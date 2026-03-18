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
	Title        string  `json:"title,omitempty"`
	SpaceID      int     `json:"space_id,omitempty"`
	SpaceName    string  `json:"space_name,omitempty"`
	InvitationID int     `json:"invitation_id,omitempty"`
}

// GetNotifications godoc
// @Summary List notifications
// @Description List notifications for the current user (overdue and due_soon). Sorted by created_at DESC.
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Max items (default 50)"
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
	// Exclude space_invitation notifications for invitations already accepted/declined
	rows, err := h.DB.Query(`
		SELECT n.id, n.todo_id, n.type, n.created_at, n.read_at, t.title, n.space_id, n.invitation_id, s.name
		FROM notifications n
		LEFT JOIN todos t ON t.id = n.todo_id
		LEFT JOIN spaces s ON s.id = n.space_id
		LEFT JOIN space_invitations i ON i.id = n.invitation_id AND n.type = 'space_invitation'
		WHERE n.user_id = $1
		AND (n.type != 'space_invitation' OR (i.id IS NOT NULL AND i.status = 'pending'))
		ORDER BY n.created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		http.Error(w, "Failed to list notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []NotificationItem
	for rows.Next() {
		var item NotificationItem
		var readAt sql.NullTime
		var createdAt sql.NullTime
		var todoID sql.NullInt64
		var title sql.NullString
		var spaceID sql.NullInt64
		var invitationID sql.NullInt64
		var spaceName sql.NullString
		if err := rows.Scan(&item.ID, &todoID, &item.Type, &createdAt, &readAt, &title, &spaceID, &invitationID, &spaceName); err != nil {
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
		list = append(list, item)
	}
	if list == nil {
		list = []NotificationItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// MarkNotificationRead godoc
// @Summary Mark notification as read
// @Description Set read_at for the notification. User must own the notification.
// @Tags notifications
// @Security BearerAuth
// @Param id path int true "Notification ID"
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
	res, err := h.DB.Exec(
		"UPDATE notifications SET read_at = NOW() WHERE id = $1 AND user_id = $2",
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
	if r.Method == http.MethodPatch || r.Method == http.MethodPut {
		h.MarkNotificationRead(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
