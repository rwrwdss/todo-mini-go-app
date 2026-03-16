package auth

import "context"

type contextKey string

const UserIDKey contextKey = "user_id"

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(UserIDKey).(int)
	return id, ok
}
