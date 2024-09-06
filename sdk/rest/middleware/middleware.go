package middleware

import (
	"github.com/chaihaobo/gocommon/logger"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type (
	Middlewares []Middleware
	Middleware  interface {
		PreRequestHook(*resty.Client, *http.Request) error
		OnAfterResponse(*resty.Client, *resty.Response) error
		OnError(*resty.Request, error)
	}
)

func MiddleWares(logger logger.Logger) Middlewares {
	return []Middleware{
		NewLoggingMiddleware(logger),
		NewTraceMiddleware(),
	}
}

func (m Middlewares) Apply(client *resty.Client) {
	var preRequestHook func(*resty.Client, *http.Request) error
	for _, middleware := range m {
		preRequestHook = combinePreRequestHook(preRequestHook, middleware.PreRequestHook)
		client.OnAfterResponse(middleware.OnAfterResponse)
		client.OnError(middleware.OnError)
	}
	client.SetPreRequestHook(preRequestHook)
}

func combinePreRequestHook(hook resty.PreRequestHook, hook2 resty.PreRequestHook) resty.PreRequestHook {
	if hook == nil {
		return hook2
	}
	return func(client *resty.Client, request *http.Request) error {
		err := hook(client, request)
		if err != nil {
			return err
		}
		return hook2(client, request)
	}
}
