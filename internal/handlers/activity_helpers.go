package handlers

import (
	"database/sql"
	"encoding/json"
)

func (h *Handler) createNotification(userID int, todoID *int, nType string, spaceID *int, invitationID *int) {
	var todoVal interface{}
	var spaceVal interface{}
	var invitationVal interface{}
	if todoID != nil {
		todoVal = *todoID
	}
	if spaceID != nil {
		spaceVal = *spaceID
	}
	if invitationID != nil {
		invitationVal = *invitationID
	}
	_, _ = h.DB.Exec(
		"INSERT INTO notifications (user_id, todo_id, type, space_id, invitation_id) VALUES ($1, $2, $3, $4, $5)",
		userID, todoVal, nType, spaceVal, invitationVal,
	)
}

func (h *Handler) logSpaceActivity(spaceID int, actorID *int, eventType string, todoID *int, subjectUserID *int, payload map[string]interface{}) {
	var actorVal interface{}
	var todoVal interface{}
	var subjectVal interface{}
	if actorID != nil {
		actorVal = *actorID
	}
	if todoID != nil {
		todoVal = *todoID
	}
	if subjectUserID != nil {
		subjectVal = *subjectUserID
	}
	var payloadVal interface{}
	if payload != nil {
		if raw, err := json.Marshal(payload); err == nil {
			payloadVal = raw
		}
	}
	_, _ = h.DB.Exec(
		`INSERT INTO space_activity_logs (space_id, actor_id, event_type, todo_id, subject_user_id, payload)
		 VALUES ($1, $2, $3, $4, $5, $6::jsonb)`,
		spaceID, actorVal, eventType, todoVal, subjectVal, payloadVal,
	)
}

func nullIntPtr(v sql.NullInt64) *int {
	if !v.Valid || v.Int64 < 1 {
		return nil
	}
	n := int(v.Int64)
	return &n
}
