package auth

import (
	"context"
	"errors"
)

type contextKey string

const userIDKey contextKey = "authenticated_user_id"

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("authenticated user is required")
	}
	return userID, nil
}
