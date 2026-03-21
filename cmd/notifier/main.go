// Notifier microservice: periodically finds todos with due_date (overdue or due soon)
// and creates notifications for assignees. Uses same DATABASE_URL as main server.
package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
	"todo-go-app/internal/config"
	"todo-go-app/internal/storage"
)

func main() {
	cfg := config.LoadNotifier()
	config.InitLog("NOTIFIER")

	db, err := storage.ConnectPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	log.Printf("database connected, check_interval=%s dedupe_hours=%d", cfg.CheckInterval, cfg.DedupeHours)

	run(db, cfg.DedupeHours)
	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()
	for range ticker.C {
		run(db, cfg.DedupeHours)
	}
}

func run(db *sql.DB, dedupeHours int) {
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	tomorrowEnd := todayStart.AddDate(0, 0, 2) // end of tomorrow
	dedupeSince := now.Add(-time.Duration(dedupeHours) * time.Hour)

	rows, err := db.Query(`
		SELECT id, COALESCE(assignee_id, user_id) as assignee_id,
			COALESCE(due_at, (due_date::text || ' 00:00:00')::timestamptz) as due_at
		FROM todos
		WHERE (due_at IS NOT NULL OR due_date IS NOT NULL) AND done = FALSE
	`)
	if err != nil {
		log.Printf("query todos: %v", err)
		return
	}
	defer rows.Close()

	var overdue, dueSoon []struct{ todoID, assigneeID int }
	for rows.Next() {
		var todoID, assigneeID int
		var dueAt sql.NullTime
		if err := rows.Scan(&todoID, &assigneeID, &dueAt); err != nil || !dueAt.Valid {
			continue
		}
		t := dueAt.Time.UTC()
		if t.Before(now) {
			overdue = append(overdue, struct{ todoID, assigneeID int }{todoID, assigneeID})
		} else if !t.Before(todayStart) && t.Before(tomorrowEnd) {
			dueSoon = append(dueSoon, struct{ todoID, assigneeID int }{todoID, assigneeID})
		}
	}

	for _, x := range overdue {
		if shouldCreate(db, x.todoID, x.assigneeID, "task_overdue", dedupeSince) {
			_, _ = db.Exec(
				"INSERT INTO notifications (user_id, todo_id, type) VALUES ($1, $2, 'task_overdue')",
				x.assigneeID, x.todoID,
			)
		}
	}
	for _, x := range dueSoon {
		if shouldCreate(db, x.todoID, x.assigneeID, "task_due_soon", dedupeSince) {
			_, _ = db.Exec(
				"INSERT INTO notifications (user_id, todo_id, type) VALUES ($1, $2, 'task_due_soon')",
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
