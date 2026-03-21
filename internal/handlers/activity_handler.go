package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"todo-go-app/internal/auth"
)

type SpaceActivityItem struct {
	ID            int64                  `json:"id"`
	SpaceID       int                    `json:"space_id"`
	ActorID       *int                   `json:"actor_id,omitempty"`
	ActorName     string                 `json:"actor_name,omitempty"`
	EventType     string                 `json:"event_type"`
	TodoID        *int                   `json:"todo_id,omitempty"`
	SubjectUserID *int                   `json:"subject_user_id,omitempty"`
	SubjectName   string                 `json:"subject_name,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	CreatedAt     string                 `json:"created_at"`
}

// GetSpaceActivity godoc
// @Summary Get space activity
// @Description Returns activity log for corporate space. Only space admins can access.
// @Tags spaces
// @Produce json
// @Security BearerAuth
// @Param id path int true "Space ID"
// @Param limit query int false "Max items (default 100)"
// @Param event_type query string false "Filter by event type"
// @Success 200 {array} SpaceActivityItem
// @Failure 403 "Forbidden"
// @Router /api/spaces/{id}/activity [get]
func (h *Handler) GetSpaceActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/spaces/")
	trimmed = strings.TrimSuffix(trimmed, "/activity")
	trimmed = strings.Trim(trimmed, "/")
	spaceID, err := strconv.Atoi(trimmed)
	if err != nil || spaceID < 1 {
		http.Error(w, "Invalid space id", http.StatusBadRequest)
		return
	}
	if !h.canAdminSpace(spaceID, userID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	limit := 100
	if q := r.URL.Query().Get("limit"); q != "" {
		if l, err := strconv.Atoi(q); err == nil && l > 0 && l <= 300 {
			limit = l
		}
	}
	eventType := strings.TrimSpace(r.URL.Query().Get("event_type"))
	query := `
		SELECT l.id, l.space_id, l.actor_id, COALESCE(au.name, au.email, ''), l.event_type, l.todo_id, l.subject_user_id, COALESCE(su.name, su.email, ''), l.payload, l.created_at
		FROM space_activity_logs l
		LEFT JOIN users au ON au.id = l.actor_id
		LEFT JOIN users su ON su.id = l.subject_user_id
		WHERE l.space_id = $1
	`
	args := []interface{}{spaceID}
	if eventType != "" {
		query += " AND l.event_type = $" + strconv.Itoa(len(args)+1)
		args = append(args, eventType)
	}
	query += " ORDER BY l.created_at DESC LIMIT $" + strconv.Itoa(len(args)+1)
	args = append(args, limit)
	rows, err := h.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to load activity", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	items := make([]SpaceActivityItem, 0, limit)
	for rows.Next() {
		var item SpaceActivityItem
		var actorID sql.NullInt64
		var todoID sql.NullInt64
		var subjectID sql.NullInt64
		var payloadRaw []byte
		var createdAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.SpaceID, &actorID, &item.ActorName, &item.EventType, &todoID, &subjectID, &item.SubjectName, &payloadRaw, &createdAt); err != nil {
			continue
		}
		if actorID.Valid {
			v := int(actorID.Int64)
			item.ActorID = &v
		}
		if todoID.Valid {
			v := int(todoID.Int64)
			item.TodoID = &v
		}
		if subjectID.Valid {
			v := int(subjectID.Int64)
			item.SubjectUserID = &v
		}
		if len(payloadRaw) > 0 {
			_ = json.Unmarshal(payloadRaw, &item.Payload)
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		items = append(items, item)
	}
	if items == nil {
		items = []SpaceActivityItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}
