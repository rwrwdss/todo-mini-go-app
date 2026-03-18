package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"todo-go-app/internal/auth"
)

// SpaceListItem for GET /api/spaces
type SpaceListItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Role string `json:"role"`
}

// SpaceDetail for GET /api/spaces/:id
type SpaceDetail struct {
	ID      int            `json:"id"`
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	OwnerID int            `json:"owner_id"`
	Members []SpaceMember  `json:"members,omitempty"`
}

// SpaceMember for space detail
type SpaceMember struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// CreateSpaceRequest for POST /api/spaces
type CreateSpaceRequest struct {
	Name string `json:"name"`
}

// InviteMemberRequest for POST /api/spaces/:id/members
type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (h *Handler) ensurePersonalSpace(userID int) (spaceID int, err error) {
	var id sql.NullInt64
	err = h.DB.QueryRow(
		"SELECT id FROM spaces WHERE owner_id = $1 AND type = 'personal'",
		userID,
	).Scan(&id)
	if err == nil && id.Valid && id.Int64 > 0 {
		return int(id.Int64), nil
	}
	err = h.DB.QueryRow(
		"INSERT INTO spaces (name, type, owner_id) VALUES ('My Space', 'personal', $1) RETURNING id",
		userID,
	).Scan(&spaceID)
	if err != nil {
		return 0, err
	}
	_, err = h.DB.Exec(
		"INSERT INTO space_members (space_id, user_id, role) VALUES ($1, $2, 'admin') ON CONFLICT (space_id, user_id) DO NOTHING",
		spaceID, userID,
	)
	if err != nil {
		return 0, err
	}
	return spaceID, nil
}

func (h *Handler) spaceRole(spaceID, userID int) (role string, ok bool) {
	err := h.DB.QueryRow(
		"SELECT role FROM space_members WHERE space_id = $1 AND user_id = $2",
		spaceID, userID,
	).Scan(&role)
	return role, err == nil
}

func (h *Handler) canAdminSpace(spaceID, userID int) bool {
	role, ok := h.spaceRole(spaceID, userID)
	return ok && (role == "admin")
}

func (h *Handler) canAccessSpace(spaceID, userID int) bool {
	_, ok := h.spaceRole(spaceID, userID)
	return ok
}

// GetSpaces godoc
// @Summary List spaces
// @Description List spaces for the current user (personal + corporate where member). Ensures personal space exists.
// @Tags spaces
// @Produce json
// @Security BearerAuth
// @Success 200 {array} SpaceListItem
// @Router /api/spaces [get]
func (h *Handler) GetSpaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	_, err := h.ensurePersonalSpace(userID)
	if err != nil {
		http.Error(w, "Failed to ensure personal space", http.StatusInternalServerError)
		return
	}
	rows, err := h.DB.Query(`
		SELECT s.id, s.name, s.type, COALESCE(sm.role, 'member')
		FROM space_members sm
		JOIN spaces s ON s.id = sm.space_id
		WHERE sm.user_id = $1
		ORDER BY s.type ASC, s.name ASC
	`, userID)
	if err != nil {
		http.Error(w, "Failed to list spaces", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var list []SpaceListItem
	for rows.Next() {
		var item SpaceListItem
		var typ string
		if err := rows.Scan(&item.ID, &item.Name, &typ, &item.Role); err != nil {
			http.Error(w, "Failed to scan space", http.StatusInternalServerError)
			return
		}
		item.Type = typ
		list = append(list, item)
	}
	if list == nil {
		list = []SpaceListItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// GetSpaceByID godoc
// @Summary Get space by ID
// @Description Get space details and members (for corporate). Caller must be a member.
// @Tags spaces
// @Produce json
// @Security BearerAuth
// @Param id path int true "Space ID"
// @Success 200 {object} SpaceDetail
// @Failure 403 "Not a member"
// @Failure 404 "Space not found"
// @Router /api/spaces/{id} [get]
func (h *Handler) GetSpaceByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/spaces/")
	idStr = strings.Trim(idStr, "/")
	spaceID, err := strconv.Atoi(idStr)
	if err != nil || spaceID < 1 {
		http.Error(w, "Invalid space id", http.StatusBadRequest)
		return
	}
	if !h.canAccessSpace(spaceID, userID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var name, typ string
	var ownerID int
	err = h.DB.QueryRow(
		"SELECT name, type, owner_id FROM spaces WHERE id = $1",
		spaceID,
	).Scan(&name, &typ, &ownerID)
	if err == sql.ErrNoRows {
		http.Error(w, "Space not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	detail := SpaceDetail{ID: spaceID, Name: name, Type: typ, OwnerID: ownerID}
	if typ == "corporate" {
		memRows, err := h.DB.Query(`
			SELECT u.id, u.email, u.name, sm.role
			FROM space_members sm
			JOIN users u ON u.id = sm.user_id
			WHERE sm.space_id = $1
			ORDER BY sm.role DESC, u.name
		`, spaceID)
		if err != nil {
			http.Error(w, "Failed to load members", http.StatusInternalServerError)
			return
		}
		defer memRows.Close()
		for memRows.Next() {
			var m SpaceMember
			if err := memRows.Scan(&m.ID, &m.Email, &m.Name, &m.Role); err != nil {
				continue
			}
			detail.Members = append(detail.Members, m)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

// CreateSpace godoc
// @Summary Create corporate space
// @Description Create a new corporate space. Caller becomes owner and admin.
// @Tags spaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateSpaceRequest true "name"
// @Success 200 {object} SpaceListItem
// @Router /api/spaces [post]
func (h *Handler) CreateSpace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req CreateSpaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	var spaceID int
	err := h.DB.QueryRow(
		"INSERT INTO spaces (name, type, owner_id) VALUES ($1, 'corporate', $2) RETURNING id",
		name, userID,
	).Scan(&spaceID)
	if err != nil {
		http.Error(w, "Failed to create space", http.StatusInternalServerError)
		return
	}
	_, err = h.DB.Exec(
		"INSERT INTO space_members (space_id, user_id, role) VALUES ($1, $2, 'admin')",
		spaceID, userID,
	)
	if err != nil {
		http.Error(w, "Failed to add member", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SpaceListItem{ID: spaceID, Name: name, Type: "corporate", Role: "admin"})
}

// InviteMember godoc
// @Summary Invite member to space
// @Description Add a user to the space by email. Caller must be admin. User must already exist.
// @Tags spaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Space ID"
// @Param body body InviteMemberRequest true "email, role"
// @Success 200 "OK"
// @Failure 403 "Not admin"
// @Failure 404 "User not found"
// @Router /api/spaces/{id}/members [post]
func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/spaces/")
	idStr = strings.Split(idStr, "/")[0]
	spaceID, err := strconv.Atoi(idStr)
	if err != nil || spaceID < 1 {
		http.Error(w, "Invalid space id", http.StatusBadRequest)
		return
	}
	if !h.canAdminSpace(spaceID, userID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var req InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "member"
	}
	if role != "admin" && role != "member" {
		role = "member"
	}
	var inviteeID int
	err = h.DB.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&inviteeID)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	_, err = h.DB.Exec(
		"INSERT INTO space_members (space_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT (space_id, user_id) DO UPDATE SET role = $3",
		spaceID, inviteeID, role,
	)
	if err != nil {
		http.Error(w, "Failed to add member", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// SpacesRouter handles /api/spaces (GET list, POST create) and delegates /api/spaces/:id to SpacesByID.
func (h *Handler) SpacesRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/spaces" || r.URL.Path == "/api/spaces/" {
		switch r.Method {
		case http.MethodGet:
			h.GetSpaces(w, r)
			return
		case http.MethodPost:
			h.CreateSpace(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.SpacesByID(w, r)
}

// SpacesByID handles GET /api/spaces/:id and POST /api/spaces/:id/members.
func (h *Handler) SpacesByID(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/spaces/")
	trimmed = strings.Trim(trimmed, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	idStr := parts[0]
	spaceID, err := strconv.Atoi(idStr)
	if err != nil || spaceID < 1 {
		http.Error(w, "Invalid space id", http.StatusBadRequest)
		return
	}
	if len(parts) == 2 && parts[1] == "members" {
		if r.Method == http.MethodPost {
			h.InviteMember(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Method == http.MethodGet {
		h.GetSpaceByID(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
