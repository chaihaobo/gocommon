package queue

import (
	"time"
)

type (
	Option interface {
		apply(*options)
	}
	OptionFunc func(*options)
	options    struct {
		delayDuration time.Duration
	}
)

func (o OptionFunc) apply(options *options) {
	o.apply(options)
}

func WithDelay(duration time.Duration) Option {
	return OptionFunc(func(o *options) {
		o.delayDuration = duration
	})
}
