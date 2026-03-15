package main

import (
	"log"
	"net/http"

	_ "todo-go-app/docs" // for swagger docs
	"todo-go-app/internal/handlers"
	"todo-go-app/internal/storage"

	httpSwagger "github.com/swaggo/http-swagger" // for swagger UI
)

func main() {

	db, err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	h := &handlers.Handler{
		DB: db,
	}

	http.HandleFunc("/api/todos", h.GetTodos)
	http.HandleFunc("/api/create", h.CreateTodo)
	http.Handle("/swagger/", httpSwagger.WrapHandler)
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	log.Println("Server started on :8080")

	http.ListenAndServe(":8080", nil)
}
