package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
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
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
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
