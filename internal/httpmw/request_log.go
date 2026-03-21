package httpmw

import (
	"log"
	"net/http"
	"time"

	"todo-go-app/internal/auth"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func LogRequests(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		userID, ok := auth.UserIDFromContext(r.Context())
		if ok {
			log.Printf("[http] %s %s status=%d dur=%s user_id=%d", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Millisecond), userID)
			return
		}
		log.Printf("[http] %s %s status=%d dur=%s", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Millisecond))
	}
}
