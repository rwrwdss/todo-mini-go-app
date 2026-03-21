package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"todo-go-app/internal/auth"
)

// InvitationItem for GET /api/invitations
type InvitationItem struct {
	ID        int    `json:"id"`
	SpaceID   int    `json:"space_id"`
	SpaceName string `json:"space_name"`
	InviterID int    `json:"inviter_id"`
	InviterName string `json:"inviter_name,omitempty"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

// GetInvitations returns pending space invitations for the current user.
func (h *Handler) GetInvitations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := h.DB.Query(`
		SELECT i.id, i.space_id, s.name, i.inviter_id, u.name, i.role, i.created_at
		FROM space_invitations i
		JOIN spaces s ON s.id = i.space_id
		JOIN users u ON u.id = i.inviter_id
		WHERE i.invitee_id = $1 AND i.status = 'pending'
		ORDER BY i.created_at DESC
	`, userID)
	if err != nil {
		http.Error(w, "Failed to list invitations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var list []InvitationItem
	for rows.Next() {
		var item InvitationItem
		var inviterName string
		var createdAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.SpaceID, &item.SpaceName, &item.InviterID, &inviterName, &item.Role, &createdAt); err != nil {
			continue
		}
		item.InviterName = inviterName
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format(time.RFC3339)
		}
		list = append(list, item)
	}
	if list == nil {
		list = []InvitationItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// AcceptInvitation accepts a space invitation (invitee = current user).
// Id can be taken from PathValue("id") when using route "POST /api/invitations/{id}/accept", or parsed from r.URL.Path.
func (h *Handler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var idStr string
	if pv := r.PathValue("id"); pv != "" {
		idStr = pv
	} else {
		idStr = strings.TrimPrefix(r.URL.Path, "/api/invitations/")
		idStr = strings.TrimSuffix(idStr, "/accept")
		idStr = strings.Trim(idStr, "/")
	}
	invitationID, err := strconv.Atoi(idStr)
	if err != nil || invitationID < 1 {
		http.Error(w, "Invalid invitation id", http.StatusBadRequest)
		return
	}
	var spaceID int
	var role string
	err = h.DB.QueryRow(
		"SELECT space_id, role FROM space_invitations WHERE id = $1 AND invitee_id = $2 AND status = 'pending'",
		invitationID, userID,
	).Scan(&spaceID, &role)
	if err != nil {
		http.Error(w, "Invitation not found or already handled", http.StatusNotFound)
		return
	}
	_, err = h.DB.Exec(
		"INSERT INTO space_members (space_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT (space_id, user_id) DO UPDATE SET role = $3",
		spaceID, userID, role,
	)
	if err != nil {
		http.Error(w, "Failed to join space", http.StatusInternalServerError)
		return
	}
	_, _ = h.DB.Exec("UPDATE space_invitations SET status = 'accepted' WHERE id = $1", invitationID)
	actorID := userID
	subjectID := userID
	h.logSpaceActivity(spaceID, &actorID, "member_joined_space", nil, &subjectID, map[string]interface{}{"role": role})
	w.WriteHeader(http.StatusOK)
}

// DeclineInvitation declines a space invitation.
// Id can be from PathValue("id") or parsed from r.URL.Path.
func (h *Handler) DeclineInvitation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var idStr string
	if pv := r.PathValue("id"); pv != "" {
		idStr = pv
	} else {
		idStr = strings.TrimPrefix(r.URL.Path, "/api/invitations/")
		idStr = strings.TrimSuffix(idStr, "/decline")
		idStr = strings.Trim(idStr, "/")
	}
	invitationID, err := strconv.Atoi(idStr)
	if err != nil || invitationID < 1 {
		http.Error(w, "Invalid invitation id", http.StatusBadRequest)
		return
	}
	res, err := h.DB.Exec(
		"UPDATE space_invitations SET status = 'declined' WHERE id = $1 AND invitee_id = $2 AND status = 'pending'",
		invitationID, userID,
	)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "Invitation not found or already handled", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// InvitationsRouter handles /api/invitations (GET) and /api/invitations/:id/accept, /api/invitations/:id/decline (POST).
func (h *Handler) InvitationsRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/invitations" || r.URL.Path == "/api/invitations/" {
		h.GetInvitations(w, r)
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/invitations/")
	trimmed = strings.Trim(trimmed, "/")
	if strings.HasSuffix(trimmed, "/accept") {
		h.AcceptInvitation(w, r)
		return
	}
	if strings.HasSuffix(trimmed, "/decline") {
		h.DeclineInvitation(w, r)
		return
	}
	http.NotFound(w, r)
}
