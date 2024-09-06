package rest

type (
	options struct {
		headers    map[string]string
		pathParams map[string]string
		debug      bool
	}

	Option interface {
		apply(*options)
	}

	optionFunc func(*options)
)

func (o optionFunc) apply(opts *options) {
	o(opts)
}

func WithHeaders(headers map[string]string) Option {
	return optionFunc(func(opts *options) { opts.headers = headers })
}

func WithDebug() Option {
	return optionFunc(func(opts *options) { opts.debug = true })
}

func WithPathParams(pathParams map[string]string) Option {
	return optionFunc(func(opts *options) { opts.pathParams = pathParams })
}
