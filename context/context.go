package context

import (
	"context"
	"time"
)

// asyncContext	Suitable for scenarios where goroutine is run asynchronously,
// it will inherit the context value of the parent context. and never Done.
//
// example:
//
//	func foo(ctx context.Context) {
//		go bar(Async(ctx))
//	}
//	func bar(ctx context.Context) {}
type asyncContext struct {
	parent context.Context
}

func (u asyncContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (u asyncContext) Done() <-chan struct{} {
	return nil
}

func (u asyncContext) Err() error {
	return nil
}

func (u asyncContext) Value(key any) any {
	return u.parent.Value(key)
}

func Async(ctx context.Context) context.Context {
	return asyncContext{parent: ctx}
}
