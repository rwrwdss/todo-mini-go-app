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
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// GetTodos godoc
// @Summary Get all todos
// @Description get list of todos
// @Tags todos
// @Produce json
// @Success 200 {array} Todo
// @Router /api/todos [get]
func (h *Handler) GetTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rows, err := h.DB.Query("SELECT id, title, done FROM todos")
	if err != nil {
		http.Error(w, "Failed to query todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Done); err != nil {
			http.Error(w, "Failed to scan todo", http.StatusInternalServerError)
			return
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
// @Description create new todo
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

	result, err := h.DB.Exec("INSERT INTO todos (title, done) VALUES (?, ?)", todo.Title, todo.Done)
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
		Title *string `json:"title"`
		Done  *bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	var title string
	var done bool
	if err := h.DB.QueryRow("SELECT title, done FROM todos WHERE id = ?", id).Scan(&title, &done); err != nil {
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
	_, err := h.DB.Exec("UPDATE todos SET title = ?, done = ? WHERE id = ?", title, done, id)
	if err != nil {
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}
	todo := Todo{ID: id, Title: title, Done: done}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

// DeleteTodo godoc
// @Summary Delete todo
// @Description delete todo by id
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
	res, err := h.DB.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
