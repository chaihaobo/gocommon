package pkg

import "context"

type (
	ActionHandler[T StateHolder[S], S comparable] interface {
		Invoke(ctx context.Context, stateHolder T) error
	}

	ActionHandlerFunc[T StateHolder[S], S comparable] func(context.Context, T) error
)

func (a ActionHandlerFunc[T, S]) Invoke(ctx context.Context, stateHolder T) error {
	return a(ctx, stateHolder)
}
