package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	DB *sql.DB
}

type Todo struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Done        bool    `json:"done"`
	Description string  `json:"description,omitempty"`
	Priority    string  `json:"priority,omitempty"`
	Tag         string  `json:"tag,omitempty"`
	ParentID    *int    `json:"parent_id,omitempty"`
}

// GetTodos godoc
// @Summary Get all todos
// @Description get list of todos. Optional query parent_id to filter by parent.
// @Tags todos
// @Produce json
// @Param parent_id query int false "Filter children of this todo id"
// @Success 200 {array} Todo
// @Router /api/todos [get]
func (h *Handler) GetTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var rows *sql.Rows
	var err error
	if parentID := r.URL.Query().Get("parent_id"); parentID != "" {
		pid, errAtoi := strconv.Atoi(parentID)
		if errAtoi != nil {
			http.Error(w, "Invalid parent_id", http.StatusBadRequest)
			return
		}
		rows, err = h.DB.Query("SELECT id, title, done, description, priority, tag, parent_id FROM todos WHERE parent_id = ?", pid)
	} else {
		rows, err = h.DB.Query("SELECT id, title, done, description, priority, tag, parent_id FROM todos")
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
		var parentID sql.NullInt64
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Done, &desc, &priority, &tag, &parentID); err != nil {
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
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to read todos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// CreateTodo godoc
// @Summary Create todo
// @Description create new todo (title required; description, priority, tag, parent_id optional)
// @Tags todos
// @Accept json
// @Produce json
// @Param todo body Todo true "Todo object"
// @Success 200 {object} Todo
// @Router /api/create [post]
func (h *Handler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	result, err := h.DB.Exec(
		"INSERT INTO todos (title, done, description, priority, tag, parent_id) VALUES (?, ?, ?, ?, ?, ?)",
		todo.Title, todo.Done, todo.Description, todo.Priority, todo.Tag, todo.ParentID,
	)
	if err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve last insert ID", http.StatusInternalServerError)
		return
	}

	todo.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

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
// @Description update todo by id (title and/or done). Use for edit and for complete/undo.
// @Tags todos
// @Accept json
// @Produce json
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
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	var title, description, priority, tag string
	var done bool
	var parentID sql.NullInt64
	if err := h.DB.QueryRow("SELECT title, done, description, priority, tag, parent_id FROM todos WHERE id = ?", id).Scan(&title, &done, &description, &priority, &tag, &parentID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get todo", http.StatusInternalServerError)
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
	_, err := h.DB.Exec(
		"UPDATE todos SET title = ?, done = ?, description = ?, priority = ?, tag = ?, parent_id = ? WHERE id = ?",
		title, done, description, priority, tag, parentIDVal, id,
	)
	if err != nil {
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

// DeleteTodo godoc
// @Summary Delete todo
// @Description delete todo by id (cascade: deletes all descendants first)
// @Tags todos
// @Param id path int true "Todo ID"
// @Success 204 "No content"
// @Failure 404 "Todo not found"
// @Router /api/todos/{id} [delete]
func (h *Handler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, ok := parseTodoID(r.URL.Path, "/api/todos/")
	if !ok {
		http.Error(w, "Invalid todo id", http.StatusBadRequest)
		return
	}
	if err := h.deleteTodoCascade(id); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteTodoCascade(id int) error {
	rows, err := h.DB.Query("SELECT id FROM todos WHERE parent_id = ?", id)
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
		if err := h.deleteTodoCascade(cid); err != nil {
			return err
		}
	}
	res, err := h.DB.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// TodosByID маршрутизирует PATCH и DELETE для /api/todos/{id}.
func (h *Handler) TodosByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPatch, http.MethodPut:
		h.UpdateTodo(w, r)
	case http.MethodDelete:
		h.DeleteTodo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
