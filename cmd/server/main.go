package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "todo-go-app/docs" // for swagger docs
	"todo-go-app/internal/handlers"
	"todo-go-app/internal/storage"

	httpSwagger "github.com/swaggo/http-swagger" // for swagger UI
)

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
	db, err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	h := &handlers.Handler{
		DB: db,
	}

	http.HandleFunc("/api/todos", h.GetTodos)
	http.HandleFunc("/api/create", h.CreateTodo)
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	webDist := "./web/dist"
	if _, err := os.Stat(webDist); err == nil {
		http.Handle("/", spaHandler(webDist))
		log.Println("Serving frontend from web/dist")
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
		log.Println("Frontend not found (web/dist). Run 'cd web && npm run build' to serve React app")
	}

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
