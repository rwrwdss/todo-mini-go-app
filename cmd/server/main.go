package main

import (
	"log"
	"net/http"

	"todo-go-app/internal/handlers"
	"todo-go-app/internal/storage"
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

	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Println("Server started on :8080")

	http.ListenAndServe(":8080", nil)
}
