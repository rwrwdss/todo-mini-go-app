package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"todo-go-app/internal/auth"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// Register godoc
// @Summary Register new user
// @Description Create account with email, password (min 8 chars), and optional name.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "email, password, name"
// @Success 200 {object} AuthResponse
// @Failure 400 "Invalid input"
// @Failure 409 "Email already registered"
// @Router /api/auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	var id int
	err = h.DB.QueryRow(
		"INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id",
		req.Email, string(hash), req.Name,
	).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(id)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  UserResponse{ID: id, Email: req.Email, Name: req.Name},
	})
}

// Login godoc
// @Summary Login
// @Description Returns JWT and user info. Use token in Authorization: Bearer <token> for /api/todos and /api/create.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "email, password"
// @Success 200 {object} AuthResponse
// @Failure 400 "Email and password required"
// @Failure 401 "Invalid email or password"
// @Router /api/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password required", http.StatusBadRequest)
		return
	}

	var id int
	var name string
	var passwordHash string
	err := h.DB.QueryRow(
		"SELECT id, name, password_hash FROM users WHERE email = $1",
		req.Email,
	).Scan(&id, &name, &passwordHash)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(id)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  UserResponse{ID: id, Email: req.Email, Name: name},
	})
}
