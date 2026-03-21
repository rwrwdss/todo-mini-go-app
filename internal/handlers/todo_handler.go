package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"todo-go-app/internal/auth"
)

type Handler struct {
	DB *sql.DB
}

type Todo struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	Done         bool    `json:"done"`
	Description  string  `json:"description,omitempty"`
	Priority     string  `json:"priority,omitempty"`
	Tag          string  `json:"tag,omitempty"`
	ParentID     *int    `json:"parent_id,omitempty"`
	SpaceID      *int    `json:"space_id,omitempty"`
	AssigneeID   *int    `json:"assignee_id,omitempty"`
	DueDate      *string `json:"due_date,omitempty"` // date only (YYYY-MM-DD), for display
	DueAt        *string `json:"due_at,omitempty"`    // full ISO datetime
	CreatedAt    *string `json:"created_at,omitempty"`
	CreatorID    *int    `json:"creator_id,omitempty"`
	CreatorName  string  `json:"creator_name,omitempty"`
	SpaceName    string  `json:"space_name,omitempty"` // filled only in from-other-spaces response
}

// GetTodos godoc
// @Summary Get todos in a space
// @Description get list of todos in the given space. Use space_id query (default: user's personal space). In corporate space, members see only their assigned tasks unless admin.
// @Tags todos
// @Produce json
// @Security BearerAuth
// @Param space_id query int false "Space ID (default: personal)"
// @Param parent_id query int false "Filter children of this todo id"
// @Success 200 {array} Todo
// @Router /api/todos [get]
func (h *Handler) GetTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	spaceID, err := h.resolveSpaceID(r, userID)
	if err != nil {
		log.Printf("[todos] resolveSpaceID userID=%d: %v", userID, err)
		if strings.Contains(err.Error(), "ensure personal space") {
			http.Error(w, "Failed to ensure personal space", http.StatusInternalServerError)
			return
		}
		http.Error(w, "Invalid space_id", http.StatusBadRequest)
		return
	}
	if !h.canAccessSpace(spaceID, userID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	// In corporate space, non-admins see only tasks assigned to them
	admin := h.canAdminSpace(spaceID, userID)

	var rows *sql.Rows
	sel := "SELECT id, title, done, description, priority, tag, parent_id, space_id, assignee_id, user_id, due_at, due_date, created_at FROM todos"
	if parentID := r.URL.Query().Get("parent_id"); parentID != "" {
		pid, errAtoi := strconv.Atoi(parentID)
		if errAtoi != nil {
			http.Error(w, "Invalid parent_id", http.StatusBadRequest)
			return
		}
		if admin {
			rows, err = h.DB.Query(sel+" WHERE space_id = $1 AND parent_id = $2", spaceID, pid)
		} else {
			rows, err = h.DB.Query(sel+" WHERE space_id = $1 AND parent_id = $2 AND assignee_id = $3", spaceID, pid, userID)
		}
	} else {
		if admin {
			rows, err = h.DB.Query(sel+" WHERE space_id = $1", spaceID)
		} else {
			rows, err = h.DB.Query(sel+" WHERE space_id = $1 AND assignee_id = $2", spaceID, userID)
		}
	}
	if err != nil {
		http.Error(w, "Failed to query todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		var desc, priority, tag sql.NullString
		var parentID, sid, assignID, creatorID sql.NullInt64
		var dueAt, dueDate, createdAt sql.NullTime
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Done, &desc, &priority, &tag, &parentID, &sid, &assignID, &creatorID, &dueAt, &dueDate, &createdAt); err != nil {
			http.Error(w, "Failed to scan todo", http.StatusInternalServerError)
			return
		}
		if desc.Valid {
			todo.Description = desc.String
		}
		if priority.Valid {
			todo.Priority = priority.String
		}
		if tag.Valid {
			todo.Tag = tag.String
		}
		if parentID.Valid && parentID.Int64 > 0 {
			p := int(parentID.Int64)
			todo.ParentID = &p
		}
		if sid.Valid && sid.Int64 > 0 {
			s := int(sid.Int64)
			todo.SpaceID = &s
		}
		if assignID.Valid && assignID.Int64 > 0 {
			a := int(assignID.Int64)
			todo.AssigneeID = &a
		}
		if creatorID.Valid && creatorID.Int64 > 0 {
			c := int(creatorID.Int64)
			todo.CreatorID = &c
		}
		if dueAt.Valid {
			todo.DueAt = strPtr(dueAt.Time.Format(time.RFC3339))
			todo.DueDate = strPtr(dueAt.Time.Format("2006-01-02"))
		} else if dueDate.Valid {
			todo.DueDate = strPtr(dueDate.Time.Format("2006-01-02"))
		}
		if createdAt.Valid {
			todo.CreatedAt = strPtr(createdAt.Time.Format(time.RFC3339))
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to read todos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func strPtr(s string) *string { return &s }

func (h *Handler) resolveSpaceID(r *http.Request, userID int) (int, error) {
	if q := r.URL.Query().Get("space_id"); q != "" {
		sid, err := strconv.Atoi(q)
		if err != nil {
			return 0, fmt.Errorf("invalid space_id: %w", err)
		}
		if sid < 1 {
			return 0, fmt.Errorf("space_id must be positive")
		}
		return sid, nil
	}
	spaceID, err := h.ensurePersonalSpace(userID)
	if err != nil {
		return 0, fmt.Errorf("ensure personal space: %w", err)
	}
	return spaceID, nil
}

func (h *Handler) todoSpaceID(todoID int) (spaceID int, err error) {
	err = h.DB.QueryRow("SELECT COALESCE(space_id, 0) FROM todos WHERE id = $1", todoID).Scan(&spaceID)
	return spaceID, err
}

// CreateTodo godoc
// @Summary Create todo
// @Description create new todo. Optional space_id (default: personal), assignee_id (required in corporate; must be space member). In personal space assignee = current user.
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param todo body Todo true "Todo (title required; space_id, assignee_id optional)"
// @Success 200 {object} Todo
// @Router /api/create [post]
func (h *Handler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	todo.Title = strings.TrimSpace(todo.Title)
	if todo.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if todo.Priority == "" {
		todo.Priority = "none"
	}
	var dueAtVal interface{}
	if todo.DueAt != nil {
		s := strings.TrimSpace(*todo.DueAt)
		if s != "" {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				if t2, err2 := time.Parse("2006-01-02T15:04:05", s); err2 == nil {
					dueAtVal = t2.UTC()
				} else if t3, err3 := time.Parse("2006-01-02", s); err3 == nil {
					dueAtVal = time.Date(t3.Year(), t3.Month(), t3.Day(), 0, 0, 0, 0, time.UTC)
				} else {
					http.Error(w, "Invalid due_at format (use ISO 8601)", http.StatusBadRequest)
					return
				}
			} else {
				dueAtVal = t.UTC()
			}
		}
	} else if todo.DueDate != nil {
		s := strings.TrimSpace(*todo.DueDate)
		if s != "" {
			if t, err := time.Parse("2006-01-02", s); err != nil {
				http.Error(w, "Invalid due_date format (use YYYY-MM-DD)", http.StatusBadRequest)
				return
			} else {
				dueAtVal = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			}
		}
	}

	spaceID := 0
	if todo.SpaceID != nil && *todo.SpaceID > 0 {
		spaceID = *todo.SpaceID
	}
	if spaceID == 0 {
		var err error
		spaceID, err = h.ensurePersonalSpace(userID)
		if err != nil {
			http.Error(w, "Failed to resolve personal space", http.StatusInternalServerError)
			return
		}
	}
	if !h.canAccessSpace(spaceID, userID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	assigneeID := userID
	if spaceID > 0 {
		var spaceType string
		if err := h.DB.QueryRow("SELECT type FROM spaces WHERE id = $1", spaceID).Scan(&spaceType); err != nil {
			http.Error(w, "Space not found", http.StatusBadRequest)
			return
		}
		if spaceType == "corporate" {
			if !h.canAdminSpace(spaceID, userID) {
				http.Error(w, "Only admins can create tasks in corporate space", http.StatusForbidden)
				return
			}
			if todo.AssigneeID != nil && *todo.AssigneeID > 0 {
				if !h.canAccessSpace(spaceID, *todo.AssigneeID) {
					http.Error(w, "Assignee must be a space member", http.StatusBadRequest)
					return
				}
				assigneeID = *todo.AssigneeID
			}
		}
	}

	var id int
	err := h.DB.QueryRow(
		"INSERT INTO todos (user_id, space_id, assignee_id, title, done, description, priority, tag, parent_id, due_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id",
		userID, spaceID, assigneeID, todo.Title, todo.Done, todo.Description, todo.Priority, todo.Tag, todo.ParentID, dueAtVal,
	).Scan(&id)
	if err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	todo.ID = id
	todo.SpaceID = &spaceID
	todo.AssigneeID = &assigneeID
	if spaceID > 0 {
		actorID := userID
		todoID := id
		subjectID := assigneeID
		h.logSpaceActivity(spaceID, &actorID, "todo_created", &todoID, &subjectID, map[string]interface{}{
			"title":    todo.Title,
			"priority": todo.Priority,
			"tag":      todo.Tag,
		})
		h.createNotification(assigneeID, &todoID, "task_created", &spaceID, nil)
		if assigneeID != userID {
			h.logSpaceActivity(spaceID, &actorID, "todo_assigned", &todoID, &subjectID, nil)
			h.createNotification(assigneeID, &todoID, "task_assigned", &spaceID, nil)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

// GetTodoByID godoc
// @Summary Get todo by ID
// @Description get full task details including creator and created_at. Caller must have access to the task's space.
// @Tags todos
// @Produce json
// @Security BearerAuth
// @Param id path int true "Todo ID"
// @Success 200 {object} Todo
// @Failure 404 "Todo not found"
// @Router /api/todos/{id} [get]
func (h *Handler) GetTodoByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, ok := parseTodoID(r.URL.Path, "/api/todos/")
	if !ok {
		http.Error(w, "Invalid todo id", http.StatusBadRequest)
		return
	}
	var todo Todo
	var desc, priority, tag sql.NullString
	var parentID, sid, assignID sql.NullInt64
	var dueDate, createdAt sql.NullTime
	var creatorID sql.NullInt64
	var creatorName sql.NullString
	var dueAt sql.NullTime
	err := h.DB.QueryRow(`
		SELECT t.id, t.title, t.done, t.description, t.priority, t.tag, t.parent_id, t.space_id, t.assignee_id, t.due_at, t.due_date, t.created_at, t.user_id, u.name
		FROM todos t
		LEFT JOIN users u ON u.id = t.user_id
		WHERE t.id = $1
	`, id).Scan(&todo.ID, &todo.Title, &todo.Done, &desc, &priority, &tag, &parentID, &sid, &assignID, &dueAt, &dueDate, &createdAt, &creatorID, &creatorName)
	if err == sql.ErrNoRows {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to get todo", http.StatusInternalServerError)
		return
	}
	todoSpaceID := 0
	if sid.Valid && sid.Int64 > 0 {
		todoSpaceID = int(sid.Int64)
		todo.SpaceID = ptr(int(sid.Int64))
	}
	if todoSpaceID > 0 {
		if !h.canAccessSpace(todoSpaceID, userID) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if !h.canAdminSpace(todoSpaceID, userID) {
			if !assignID.Valid || int(assignID.Int64) != userID {
				http.Error(w, "Todo not found", http.StatusNotFound)
				return
			}
		}
	} else {
		if creatorID.Valid && int(creatorID.Int64) != userID {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
	}
	if desc.Valid {
		todo.Description = desc.String
	}
	if priority.Valid {
		todo.Priority = priority.String
	}
	if tag.Valid {
		todo.Tag = tag.String
	}
	if parentID.Valid && parentID.Int64 > 0 {
		p := int(parentID.Int64)
		todo.ParentID = &p
	}
	if assignID.Valid && assignID.Int64 > 0 {
		a := int(assignID.Int64)
		todo.AssigneeID = &a
	}
	if dueAt.Valid {
		todo.DueAt = strPtr(dueAt.Time.Format(time.RFC3339))
		todo.DueDate = strPtr(dueAt.Time.Format("2006-01-02"))
	} else if dueDate.Valid {
		todo.DueDate = strPtr(dueDate.Time.Format("2006-01-02"))
	}
	if createdAt.Valid {
		todo.CreatedAt = strPtr(createdAt.Time.Format(time.RFC3339))
	}
	if creatorID.Valid {
		c := int(creatorID.Int64)
		todo.CreatorID = &c
	}
	if creatorName.Valid {
		todo.CreatorName = creatorName.String
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func ptr(i int) *int { return &i }

// parseTodoID извлекает id из пути вида /api/todos/123.
func parseTodoID(path, prefix string) (int, bool) {
	s := strings.TrimPrefix(path, prefix)
	s = strings.Trim(s, "/")
	if s == "" {
		return 0, false
	}
	id, err := strconv.Atoi(s)
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// UpdateTodo godoc
// @Summary Update todo
// @Description update todo by id (only own todos). Use for edit and for complete/undo.
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Todo ID"
// @Param body body Todo true "Fields to update (title and/or done)"
// @Success 200 {object} Todo
// @Failure 404 "Todo not found"
// @Router /api/todos/{id} [patch]
func (h *Handler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, ok := parseTodoID(r.URL.Path, "/api/todos/")
	if !ok {
		http.Error(w, "Invalid todo id", http.StatusBadRequest)
		return
	}
	var body struct {
		Title       *string `json:"title"`
		Done        *bool   `json:"done"`
		Description *string `json:"description"`
		Priority    *string `json:"priority"`
		Tag         *string `json:"tag"`
		ParentID    *int    `json:"parent_id"`
		DueDate     *string `json:"due_date"`
		DueAt       *string `json:"due_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	var title, description, priority, tag string
	var done bool
	var parentID sql.NullInt64
	var todoSpaceID, todoUserID int
	var assigneeID sql.NullInt64
	var currentDueAt sql.NullTime
	if err := h.DB.QueryRow(
		"SELECT title, done, description, priority, tag, parent_id, COALESCE(space_id, 0), user_id, assignee_id, due_at FROM todos WHERE id = $1",
		id,
	).Scan(&title, &done, &description, &priority, &tag, &parentID, &todoSpaceID, &todoUserID, &assigneeID, &currentDueAt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get todo", http.StatusInternalServerError)
		return
	}
	prevDone := done
	prevTitle := title
	prevDescription := description
	prevPriority := priority
	prevTag := tag
	prevParentID := parentID
	prevDueAt := currentDueAt
	if todoSpaceID > 0 {
		if !h.canAccessSpace(todoSpaceID, userID) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if !h.canAdminSpace(todoSpaceID, userID) {
			if !assigneeID.Valid || int(assigneeID.Int64) != userID {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
	} else if todoUserID != userID {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	// Member (assignee, not admin, not creator) may only update done
	memberOnlyDone := todoSpaceID > 0 && !h.canAdminSpace(todoSpaceID, userID) && todoUserID != userID
	if memberOnlyDone {
		if body.Done != nil {
			done = *body.Done
		}
		_, err := h.DB.Exec("UPDATE todos SET done = $1 WHERE id = $2", done, id)
		if err != nil {
			http.Error(w, "Failed to update todo", http.StatusInternalServerError)
			return
		}
		if todoSpaceID > 0 && prevDone != done {
			actorID := userID
			todoID := id
			h.logSpaceActivity(todoSpaceID, &actorID, "todo_done_toggled", &todoID, nullIntPtr(assigneeID), map[string]interface{}{"done": done})
		}
		todo := Todo{ID: id, Title: title, Done: done, Description: description, Priority: priority, Tag: tag}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todo)
		return
	}
	if body.Title != nil {
		title = strings.TrimSpace(*body.Title)
		if title == "" {
			http.Error(w, "Title cannot be empty", http.StatusBadRequest)
			return
		}
	}
	if body.Done != nil {
		done = *body.Done
	}
	if body.Description != nil {
		description = *body.Description
	}
	if body.Priority != nil {
		priority = *body.Priority
	}
	if body.Tag != nil {
		tag = *body.Tag
	}
	var parentIDVal interface{}
	if body.ParentID != nil {
		parentIDVal = *body.ParentID
	} else if parentID.Valid {
		parentIDVal = parentID.Int64
	} else {
		parentIDVal = nil
	}
	var dueAtVal interface{}
	if body.DueAt != nil {
		s := strings.TrimSpace(*body.DueAt)
		if s == "" {
			dueAtVal = nil
		} else {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				if t2, err2 := time.Parse("2006-01-02T15:04:05", s); err2 == nil {
					dueAtVal = t2.UTC()
				} else if t3, err3 := time.Parse("2006-01-02", s); err3 == nil {
					dueAtVal = time.Date(t3.Year(), t3.Month(), t3.Day(), 0, 0, 0, 0, time.UTC)
				} else {
					http.Error(w, "Invalid due_at format (use ISO 8601)", http.StatusBadRequest)
					return
				}
			} else {
				dueAtVal = t.UTC()
			}
		}
	} else if body.DueDate != nil {
		s := strings.TrimSpace(*body.DueDate)
		if s == "" {
			dueAtVal = nil
		} else {
			if t, err := time.Parse("2006-01-02", s); err != nil {
				http.Error(w, "Invalid due_date format (use YYYY-MM-DD)", http.StatusBadRequest)
				return
			} else {
				dueAtVal = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			}
		}
	} else {
		if currentDueAt.Valid {
			dueAtVal = currentDueAt.Time
		} else {
			dueAtVal = nil
		}
	}
	_, err := h.DB.Exec(
		"UPDATE todos SET title = $1, done = $2, description = $3, priority = $4, tag = $5, parent_id = $6, due_at = $7 WHERE id = $8",
		title, done, description, priority, tag, parentIDVal, dueAtVal, id,
	)
	if err != nil {
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}
	if todoSpaceID > 0 {
		actorID := userID
		todoID := id
		changed := map[string]interface{}{}
		if prevTitle != title {
			changed["title"] = true
		}
		if prevDescription != description {
			changed["description"] = true
		}
		if prevPriority != priority {
			changed["priority"] = true
		}
		if prevTag != tag {
			changed["tag"] = true
		}
		if prevParentID.Valid != parentID.Valid || (prevParentID.Valid && parentID.Valid && prevParentID.Int64 != parentID.Int64) {
			changed["parent_id"] = true
		}
		if prevDueAt.Valid != currentDueAt.Valid || (prevDueAt.Valid && currentDueAt.Valid && !prevDueAt.Time.Equal(currentDueAt.Time)) || body.DueAt != nil || body.DueDate != nil {
			changed["due_at"] = true
		}
		if prevDone != done {
			h.logSpaceActivity(todoSpaceID, &actorID, "todo_done_toggled", &todoID, nullIntPtr(assigneeID), map[string]interface{}{"done": done})
		}
		if len(changed) > 0 {
			h.logSpaceActivity(todoSpaceID, &actorID, "todo_updated", &todoID, nullIntPtr(assigneeID), map[string]interface{}{"fields": changed})
		}
	}
	todo := Todo{ID: id, Title: title, Done: done, Description: description, Priority: priority, Tag: tag}
	if parentIDVal != nil {
		if p, ok := parentIDVal.(int); ok {
			todo.ParentID = &p
		} else if p64, ok := parentIDVal.(int64); ok {
			p := int(p64)
			todo.ParentID = &p
		}
	}
	if dueAtVal != nil {
		if t, ok := dueAtVal.(time.Time); ok {
			todo.DueAt = strPtr(t.Format(time.RFC3339))
			todo.DueDate = strPtr(t.Format("2006-01-02"))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

// DeleteTodo godoc
// @Summary Delete todo
// @Description delete todo by id (only own todos; cascade: deletes all descendants first)
// @Tags todos
// @Security BearerAuth
// @Param id path int true "Todo ID"
// @Success 204 "No content"
// @Failure 404 "Todo not found"
// @Router /api/todos/{id} [delete]
func (h *Handler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, ok := parseTodoID(r.URL.Path, "/api/todos/")
	if !ok {
		http.Error(w, "Invalid todo id", http.StatusBadRequest)
		return
	}
	spaceID, err := h.todoSpaceID(id)
	if err != nil || spaceID < 1 {
		var todoUserID int
		_ = h.DB.QueryRow("SELECT user_id FROM todos WHERE id = $1", id).Scan(&todoUserID)
		if todoUserID != userID {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
	} else {
		if !h.canAccessSpace(spaceID, userID) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		var assigneeID sql.NullInt64
		_ = h.DB.QueryRow("SELECT assignee_id FROM todos WHERE id = $1", id).Scan(&assigneeID)
		if !h.canAdminSpace(spaceID, userID) && (!assigneeID.Valid || int(assigneeID.Int64) != userID) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}
	if err := h.deleteTodoCascade(r.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteTodoCascade(ctx context.Context, id int) error {
	rows, err := h.DB.QueryContext(ctx, "SELECT id FROM todos WHERE parent_id = $1", id)
	if err != nil {
		return err
	}
	var childIDs []int
	for rows.Next() {
		var cid int
		if err := rows.Scan(&cid); err != nil {
			rows.Close()
			return err
		}
		childIDs = append(childIDs, cid)
	}
	rows.Close()
	for _, cid := range childIDs {
		if err := h.deleteTodoCascade(ctx, cid); err != nil {
			return err
		}
	}
	res, err := h.DB.Exec("DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetTodosFromOtherSpaces returns todos from spaces where the user is a member, excluding their personal space.
func (h *Handler) GetTodosFromOtherSpaces(w http.ResponseWriter, r *http.Request) {
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
		SELECT t.id, t.title, t.done, t.description, t.priority, t.tag, t.parent_id, t.space_id, t.assignee_id, t.user_id, t.due_at, t.created_at, s.name
		FROM todos t
		JOIN space_members sm ON sm.space_id = t.space_id AND sm.user_id = $1
		JOIN spaces s ON s.id = t.space_id
		WHERE (s.owner_id != $1 OR s.type != 'personal')
		AND (sm.role = 'admin' OR t.assignee_id = $1)
		ORDER BY t.space_id, t.created_at
	`, userID)
	if err != nil {
		log.Printf("[from-other-spaces] query error userID=%d: %v", userID, err)
		http.Error(w, "Failed to list todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var todos []Todo
	for rows.Next() {
		var todo Todo
		var desc, priority, tag sql.NullString
		var parentID, sid, assignID, creatorID sql.NullInt64
		var dueAt, createdAt sql.NullTime
		var spaceName sql.NullString
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Done, &desc, &priority, &tag, &parentID, &sid, &assignID, &creatorID, &dueAt, &createdAt, &spaceName); err != nil {
			continue
		}
		if desc.Valid {
			todo.Description = desc.String
		}
		if priority.Valid {
			todo.Priority = priority.String
		}
		if tag.Valid {
			todo.Tag = tag.String
		}
		if parentID.Valid && parentID.Int64 > 0 {
			p := int(parentID.Int64)
			todo.ParentID = &p
		}
		if sid.Valid && sid.Int64 > 0 {
			s := int(sid.Int64)
			todo.SpaceID = &s
		}
		if assignID.Valid && assignID.Int64 > 0 {
			a := int(assignID.Int64)
			todo.AssigneeID = &a
		}
		if creatorID.Valid && creatorID.Int64 > 0 {
			c := int(creatorID.Int64)
			todo.CreatorID = &c
		}
		if dueAt.Valid {
			todo.DueAt = strPtr(dueAt.Time.Format(time.RFC3339))
			todo.DueDate = strPtr(dueAt.Time.Format("2006-01-02"))
		}
		if createdAt.Valid {
			todo.CreatedAt = strPtr(createdAt.Time.Format(time.RFC3339))
		}
		if spaceName.Valid {
			todo.SpaceName = spaceName.String
		}
		todos = append(todos, todo)
	}
	if todos == nil {
		todos = []Todo{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// TodosByID маршрутизирует GET, PATCH и DELETE для /api/todos/{id}.
func (h *Handler) TodosByID(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/todos/")
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "from-other-spaces" {
		h.GetTodosFromOtherSpaces(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.GetTodoByID(w, r)
	case http.MethodPatch, http.MethodPut:
		h.UpdateTodo(w, r)
	case http.MethodDelete:
		h.DeleteTodo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
