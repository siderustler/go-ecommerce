package service_logger

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	clientIDKey  contextKey = "client_id"
	userIDKey    contextKey = "user_id"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func WithClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientIDKey, clientID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func LoggerWithContext(base *slog.Logger, ctx context.Context) *slog.Logger {
	if base == nil {
		base = slog.Default()
	}

	logger := base
	if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != "" {
		logger = logger.With("request_id", requestID)
	}
	if clientID, ok := ctx.Value(clientIDKey).(string); ok && clientID != "" {
		logger = logger.With("client_id", clientID)
	}
	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		logger = logger.With("user_id", userID)
	}

	return logger
}
