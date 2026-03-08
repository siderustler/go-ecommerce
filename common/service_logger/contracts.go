package service_logger

import "context"

type Command[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

type CommandResult[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

type Query[Q, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}
