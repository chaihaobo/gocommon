package pkg

type (
	Transition[T StateHolder[S], S comparable] struct {
		from, to, failed S
		handler          ActionHandler[T, S]
		successHooks     []ActionHandler[T, S]
	}

	TransitionBuilder[T StateHolder[S], S comparable] struct {
		transition *Transition[T, S]
	}
)

func NewTransitionBuilder[T StateHolder[S], S comparable]() *TransitionBuilder[T, S] {
	return &TransitionBuilder[T, S]{
		transition: &Transition[T, S]{},
	}
}

func (t *TransitionBuilder[T, S]) To(to S) *TransitionBuilder[T, S] {
	t.transition.to = to
	return t
}

func (t *TransitionBuilder[T, S]) From(from S) *TransitionBuilder[T, S] {
	t.transition.from = from
	return t
}

func (t *TransitionBuilder[T, S]) Failed(failed S) *TransitionBuilder[T, S] {
	t.transition.failed = failed
	return t
}

func (t *TransitionBuilder[T, S]) Handler(handler ActionHandler[T, S]) *TransitionBuilder[T, S] {
	t.transition.handler = handler
	return t
}

func (t *TransitionBuilder[T, S]) SuccessHook(handler ActionHandler[T, S]) *TransitionBuilder[T, S] {
	t.transition.successHooks = append(t.transition.successHooks, handler)
	return t
}

func (t *TransitionBuilder[T, S]) Build() *Transition[T, S] {
	return t.transition
}
