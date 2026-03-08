package service_logger

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type QueryLoggerDecorator[Q, R any] struct {
	handler func(ctx context.Context, query Q) (R, error)
	logger  *slog.Logger
}

func NewQueryLoggerDecorator[Q, R any](handler func(ctx context.Context, query Q) (R, error), logger *slog.Logger) QueryLoggerDecorator[Q, R] {
	if logger == nil {
		logger = slog.Default()
	}
	return QueryLoggerDecorator[Q, R]{
		handler: handler,
		logger:  logger,
	}
}

func (d QueryLoggerDecorator[Q, R]) Handle(ctx context.Context, query Q) (result R, err error) {
	logger := LoggerWithContext(d.logger, ctx)
	startedAt := time.Now()
	logger.DebugContext(ctx, fmt.Sprintf("Handling %T query", query), "input", query)
	defer func() {
		durationMs := time.Since(startedAt).Milliseconds()
		if err != nil {
			logger.ErrorContext(ctx, fmt.Sprintf("Failed to handle %T query", query), "input", query, "error", err, "duration_ms", durationMs)
		} else {
			logger.DebugContext(ctx, fmt.Sprintf("Successfully handled %T query", query), "input", query, "result", result, "duration_ms", durationMs)
		}
	}()
	return d.handler(ctx, query)
}
