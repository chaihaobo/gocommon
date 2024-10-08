package pkg

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrActionNotDefine = errors.New("action not defined")
)

type (
	// StateMachine 状态机 执行一个动作. 完成一个状态到另外一个状态的转换
	// T StateHolder[S] T:状态持有者 S:持有的状态
	StateMachine[T StateHolder[S], S comparable] struct {
		stateHolder T
		transitions map[string]*Transition[T, S]
	}

	StateHolder[S comparable] interface {
		State() S
		UpdateState(S, error)
	}
)

// AddTransition 添加状态转换的映射. 定义了一条规则: 执行一个动作需要的状态为from, 执行动作为action, 执行动作成功后状态为to, 执行动作失败后状态为failed
// action 执行的动作
// from 执行动作前的状态
// to 执行动作后的状态
// failed 执行动作失败后的状态
// handler 执行动作的处理函数
// afterHooks 执行动作成功后需要执行的钩子函数
func (s *StateMachine[T, S]) AddTransition(action string, transition *Transition[T, S]) *StateMachine[T, S] {
	s.transitions[action] = transition
	return s
}

func NewStateMachine[T StateHolder[S], S comparable](stateHolder T) StateMachine[T, S] {
	return StateMachine[T, S]{
		stateHolder: stateHolder,
		transitions: make(map[string]*Transition[T, S]),
	}
}

// Submit 执行一个动作。完成一个状态到另外一个状态的转换
func (s *StateMachine[T, S]) Submit(ctx context.Context, action string) error {
	transition, ok := s.transitions[action]
	if !ok {
		return ErrActionNotDefine
	}
	fromState := transition.from
	if currentState := s.stateHolder.State(); fromState != currentState {
		return fmt.Errorf("current state is %v, can not execute this action %s", currentState, action)
	}
	err := transition.getHandler().Invoke(ctx, s.stateHolder)
	if err != nil {
		s.stateHolder.UpdateState(transition.failed, err)
		if err := s.triggerAfterHook(ctx, transition); err != nil {
			return err
		}
		return fmt.Errorf("failed to invoke %s action handler during state transition: %w", action, err)
	}
	s.stateHolder.UpdateState(transition.to, nil)
	return s.triggerAfterHook(ctx, transition)
}

func (s *StateMachine[T, S]) triggerAfterHook(ctx context.Context, transition *Transition[T, S]) error {
	for _, hook := range transition.afterHooks {
		err := hook.Invoke(ctx, s.stateHolder)
		if err != nil {
			return fmt.Errorf("failed to invoke success hook: %w", err)
		}
	}
	return nil
}

// Submits 执行多个连续的动作
func (s *StateMachine[T, S]) Submits(ctx context.Context, actions ...string) error {
	for _, action := range actions {
		if err := s.Submit(ctx, action); err != nil {
			return err
		}
	}
	return nil
}
