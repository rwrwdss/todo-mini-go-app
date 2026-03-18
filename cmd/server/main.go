// @title Task.grid API
// @version 1.0
// @description API for tasks with JWT auth and spaces (personal + corporate). Register or login at /api/auth/*; use GET /api/auth/check to validate session. Send token as Authorization: Bearer &lt;token&gt; for /api/todos, /api/create, /api/spaces. Todos support space_id and assignee_id for corporate workspaces.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "todo-go-app/docs" // for swagger docs
	"todo-go-app/internal/auth"
	"todo-go-app/internal/config"
	"todo-go-app/internal/handlers"
	"todo-go-app/internal/storage"

	httpSwagger "github.com/swaggo/http-swagger" // for swagger UI
)

//go:embed static/swagger-theme.css
var swaggerThemeCSS embed.FS

// spaHandler serves static files from dir and falls back to index.html for SPA routes.
func spaHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

func main() {
	cfg := config.Load()
	config.InitLog("APP")

	auth.InitJWT([]byte(cfg.JWTSecret))

	db, err := storage.InitPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	log.Printf("database connected")

	h := &handlers.Handler{
		DB: db,
	}

	http.HandleFunc("/api/auth/register", h.Register)
	http.HandleFunc("/api/auth/login", h.Login)
	http.HandleFunc("/api/auth/check", auth.RequireAuth(h.CheckSession))
	http.HandleFunc("/api/todos", auth.RequireAuth(h.GetTodos))
	http.HandleFunc("/api/todos/", auth.RequireAuth(h.TodosByID))
	http.HandleFunc("/api/create", auth.RequireAuth(h.CreateTodo))
	http.HandleFunc("/api/spaces", auth.RequireAuth(h.SpacesRouter))
	http.HandleFunc("/api/spaces/", auth.RequireAuth(h.SpacesRouter))
	http.HandleFunc("/api/invitations", auth.RequireAuth(h.InvitationsRouter))
	http.HandleFunc("/api/invitations/", auth.RequireAuth(h.InvitationsRouter))
	http.HandleFunc("POST /api/invitations/{id}/accept", auth.RequireAuth(h.AcceptInvitation))
	http.HandleFunc("POST /api/invitations/{id}/decline", auth.RequireAuth(h.DeclineInvitation))
	http.HandleFunc("/api/notifications", auth.RequireAuth(h.NotificationsRouter))
	http.HandleFunc("/api/notifications/", auth.RequireAuth(h.NotificationsRouter))
	cssFS, _ := fs.Sub(swaggerThemeCSS, "static")
	http.HandleFunc("/swagger/theme.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		http.ServeFileFS(w, r, cssFS, "swagger-theme.css")
	})
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.UIConfig(map[string]string{
			"customCssUrl": `"/swagger/theme.css"`,
		}),
	))

	webDist := cfg.WebDist
	if _, err := os.Stat(webDist); err == nil {
		http.Handle("/", spaHandler(webDist))
		log.Printf("frontend: serving from %s", webDist)
	} else {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Frontend not built. Run: cd web && npm run build\n"))
		})
		log.Printf("frontend: not built (missing %s), run 'cd web && npm run build'", webDist)
	}

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server: %v", err)
	}
}
