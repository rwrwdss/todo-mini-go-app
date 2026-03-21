package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"todo-go-app/internal/auth"
)

// Notification item for API response
type NotificationItem struct {
	ID           int     `json:"id"`
	TodoID       int     `json:"todo_id,omitempty"`
	Type         string  `json:"type"`
	CreatedAt    string  `json:"created_at"`
	ReadAt       *string `json:"read_at,omitempty"`
	ArchivedAt   *string `json:"archived_at,omitempty"`
	Title        string  `json:"title,omitempty"`
	SpaceID      int     `json:"space_id,omitempty"`
	SpaceName    string  `json:"space_name,omitempty"`
	InvitationID int     `json:"invitation_id,omitempty"`
}

// GetNotifications godoc
// @Summary List notifications
// @Description List notifications for the current user. Sorted by created_at DESC.
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Max items (default 50)"
// @Param type query string false "Filter by notification type"
// @Param include_archived query bool false "Include archived notifications (default false)"
// @Success 200 {array} NotificationItem
// @Router /api/notifications [get]
func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		if l, err := strconv.Atoi(q); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	includeArchived := false
	if q := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("include_archived"))); q == "1" || q == "true" || q == "yes" {
		includeArchived = true
	}
	typeFilter := strings.TrimSpace(r.URL.Query().Get("type"))
	// Exclude invitation notifications for already handled invitations
	query := `
		SELECT n.id, n.todo_id, n.type, n.created_at, n.read_at, n.archived_at, t.title, n.space_id, n.invitation_id, s.name
		FROM notifications n
		LEFT JOIN todos t ON t.id = n.todo_id
		LEFT JOIN spaces s ON s.id = n.space_id
		LEFT JOIN space_invitations i ON i.id = n.invitation_id AND n.type = 'space_invitation'
		WHERE n.user_id = $1
		AND (n.type != 'space_invitation' OR (i.id IS NOT NULL AND i.status = 'pending'))
	`
	args := []interface{}{userID}
	if !includeArchived {
		query += " AND n.archived_at IS NULL"
	}
	if typeFilter != "" {
		query += " AND n.type = $" + strconv.Itoa(len(args)+1)
		args = append(args, typeFilter)
	}
	query += " ORDER BY n.created_at DESC LIMIT $" + strconv.Itoa(len(args)+1)
	args = append(args, limit)
	rows, err := h.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to list notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []NotificationItem
	for rows.Next() {
		var item NotificationItem
		var readAt sql.NullTime
		var archivedAt sql.NullTime
		var createdAt sql.NullTime
		var todoID sql.NullInt64
		var title sql.NullString
		var spaceID sql.NullInt64
		var invitationID sql.NullInt64
		var spaceName sql.NullString
		if err := rows.Scan(&item.ID, &todoID, &item.Type, &createdAt, &readAt, &archivedAt, &title, &spaceID, &invitationID, &spaceName); err != nil {
			continue
		}
		if todoID.Valid {
			item.TodoID = int(todoID.Int64)
		}
		if title.Valid {
			item.Title = title.String
		}
		if spaceID.Valid {
			item.SpaceID = int(spaceID.Int64)
		}
		if invitationID.Valid {
			item.InvitationID = int(invitationID.Int64)
		}
		if spaceName.Valid {
			item.SpaceName = spaceName.String
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		if readAt.Valid {
			s := readAt.Time.Format("2006-01-02T15:04:05Z07:00")
			item.ReadAt = &s
		}
		if archivedAt.Valid {
			s := archivedAt.Time.Format("2006-01-02T15:04:05Z07:00")
			item.ArchivedAt = &s
		}
		list = append(list, item)
	}
	if list == nil {
		list = []NotificationItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// MarkNotificationRead godoc
// @Summary Update notification state
// @Description Set read/unread and archive/unarchive for the notification. User must own the notification.
// @Tags notifications
// @Accept json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Param body body map[string]bool false "read, archived"
// @Success 200 "OK"
// @Failure 404 "Not found"
// @Router /api/notifications/{id} [patch]
func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/notifications/")
	idStr = strings.Trim(idStr, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.Error(w, "Invalid notification id", http.StatusBadRequest)
		return
	}
	var body struct {
		Read     *bool `json:"read"`
		Archived *bool `json:"archived"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	setRead := body.Read == nil || *body.Read
	if body.Read != nil && !*body.Read {
		setRead = false
	}
	setArchived := false
	if body.Archived != nil {
		setArchived = *body.Archived
	}
	setReadExpr := "read_at = NOW()"
	if !setRead {
		setReadExpr = "read_at = NULL"
	}
	setArchiveExpr := "archived_at = archived_at"
	if body.Archived != nil {
		if setArchived {
			setArchiveExpr = "archived_at = NOW()"
		} else {
			setArchiveExpr = "archived_at = NULL"
		}
	}
	res, err := h.DB.Exec(
		"UPDATE notifications SET "+setReadExpr+", "+setArchiveExpr+" WHERE id = $1 AND user_id = $2",
		id, userID,
	)
	if err != nil {
		http.Error(w, "Failed to update notification", http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ArchiveNotification godoc
// @Summary Archive notification
// @Description Archive notification. User must own the notification.
// @Tags notifications
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 "OK"
// @Failure 404 "Not found"
// @Router /api/notifications/{id}/archive [post]
func (h *Handler) ArchiveNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/notifications/")
	trimmed = strings.TrimSuffix(trimmed, "/archive")
	trimmed = strings.Trim(trimmed, "/")
	id, err := strconv.Atoi(trimmed)
	if err != nil || id < 1 {
		http.Error(w, "Invalid notification id", http.StatusBadRequest)
		return
	}
	res, err := h.DB.Exec("UPDATE notifications SET archived_at = NOW() WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "Failed to archive notification", http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// NotificationsRouter handles GET /api/notifications and PATCH /api/notifications/:id
func (h *Handler) NotificationsRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/notifications" || r.URL.Path == "/api/notifications/" {
		if r.Method == http.MethodGet {
			h.GetNotifications(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if strings.HasSuffix(strings.Trim(r.URL.Path, "/"), "archive") {
		h.ArchiveNotification(w, r)
		return
	}
	if r.Method == http.MethodPatch || r.Method == http.MethodPut {
		h.MarkNotificationRead(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
