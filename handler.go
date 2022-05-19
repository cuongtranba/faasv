package faasv

import "context"

type HandlerFun[T, V any] func(ctx context.Context, arg T) (V, error)

type Handler[T, V any] interface {
	Subject() string
	Queue() string
	Handler() HandlerFun[T, V]
}

type GetUserRequest struct {
	Id string
}

type GetUserResponse struct {
	Name string
}

var _ Handler = (*ExampleHandler[GetUserRequest, GetUserResponse])(nil)

type ExampleHandler[T, V any] struct {
}

// Handler implements Handler
func (*ExampleHandler[T, V]) Handler() HandlerFun[T, V] {
	return func(ctx context.Context, arg T) (V, error) {
		return "", nil
	}
}

// Queue implements Handler
func (*ExampleHandler[T, V]) Queue() string {
	panic("unimplemented")
}

// Subject implements Handler
func (*ExampleHandler[T, V]) Subject() string {
	panic("unimplemented")
}
