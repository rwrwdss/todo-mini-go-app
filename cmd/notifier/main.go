// Notifier microservice: periodically finds todos with due_date (overdue or due soon)
// and creates notifications for assignees. Run with same DATABASE_URL as main server.
package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
	"todo-go-app/internal/storage"
)

const (
	checkInterval = 10 * time.Minute
	dedupeHours   = 24
)

func main() {
	db, err := storage.InitPostgres()
	if err != nil {
		log.Fatal("notifier: db init: ", err)
	}
	defer db.Close()

	log.Println("notifier: started, interval ", checkInterval)
	run(db)
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	for range ticker.C {
		run(db)
	}
}

func run(db *sql.DB) {
	now := time.Now().UTC()
	today := now.Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	dedupeSince := now.Add(-dedupeHours * time.Hour)

	rows, err := db.Query(`
		SELECT id, COALESCE(assignee_id, user_id) as assignee_id, due_date::text
		FROM todos
		WHERE due_date IS NOT NULL AND done = FALSE
	`)
	if err != nil {
		log.Println("notifier: query todos: ", err)
		return
	}
	defer rows.Close()

	var overdue, dueSoon []struct{ todoID, assigneeID int }
	for rows.Next() {
		var todoID, assigneeID int
		var dueStr string
		if err := rows.Scan(&todoID, &assigneeID, &dueStr); err != nil {
			continue
		}
		if dueStr < today {
			overdue = append(overdue, struct{ todoID, assigneeID int }{todoID, assigneeID})
		} else if dueStr == today || dueStr == tomorrow {
			dueSoon = append(dueSoon, struct{ todoID, assigneeID int }{todoID, assigneeID})
		}
	}

	for _, x := range overdue {
		if shouldCreate(db, x.todoID, x.assigneeID, "overdue", dedupeSince) {
			_, _ = db.Exec(
				"INSERT INTO notifications (user_id, todo_id, type) VALUES ($1, $2, 'overdue')",
				x.assigneeID, x.todoID,
			)
		}
	}
	for _, x := range dueSoon {
		if shouldCreate(db, x.todoID, x.assigneeID, "due_soon", dedupeSince) {
			_, _ = db.Exec(
				"INSERT INTO notifications (user_id, todo_id, type) VALUES ($1, $2, 'due_soon')",
				x.assigneeID, x.todoID,
			)
		}
	}
}

func shouldCreate(db *sql.DB, todoID, userID int, ntype string, since time.Time) bool {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM notifications
		WHERE todo_id = $1 AND user_id = $2 AND type = $3 AND created_at >= $4
	`, todoID, userID, ntype, since).Scan(&count)
	if err != nil {
		return true
	}
	return count == 0
}
