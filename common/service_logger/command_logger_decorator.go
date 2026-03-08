package service_logger

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type CommandLoggerDecorator[C any] struct {
	handler func(ctx context.Context, cmd C) error
	logger  *slog.Logger
}

func NewCommandLoggerDecorator[C any](handler func(ctx context.Context, cmd C) error, logger *slog.Logger) CommandLoggerDecorator[C] {
	if logger == nil {
		logger = slog.Default()
	}
	return CommandLoggerDecorator[C]{
		handler: handler,
		logger:  logger,
	}
}

func (d CommandLoggerDecorator[C]) Handle(ctx context.Context, cmd C) (err error) {
	logger := LoggerWithContext(d.logger, ctx)
	startedAt := time.Now()
	logger.DebugContext(ctx, fmt.Sprintf("Handling %T command", cmd), "input", cmd)
	defer func() {
		durationMs := time.Since(startedAt).Milliseconds()
		if err != nil {
			logger.ErrorContext(ctx, fmt.Sprintf("Failed to handle %T command", cmd), "input", cmd, "error", err, "duration_ms", durationMs)
		} else {
			logger.DebugContext(ctx, fmt.Sprintf("Successfully handled %T command", cmd), "input", cmd, "duration_ms", durationMs)
		}
	}()
	return d.handler(ctx, cmd)
}
