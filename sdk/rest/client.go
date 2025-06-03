package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/chaihaobo/gocommon/logger"
	"github.com/chaihaobo/gocommon/sdk/rest/middleware"
)

type (
	Client interface {
		//	Get 发起 Get 请求
		Get(ctx context.Context, url string, body, result any, opts ...Option) error
		//	Post 发起 Post 请求
		Post(ctx context.Context, url string, body, result any, opts ...Option) error
		//	Delete 发起 Delete 请求
		Delete(ctx context.Context, url string, body, result any, opts ...Option) error
		//	Patch 发起 Patch 请求
		Patch(ctx context.Context, url string, body, result any, opts ...Option) error
		// Execute 发起自定义请求
		Execute(ctx context.Context, method, url string, body, result any, opts ...Option) error
	}

	client struct {
		restyClient *resty.Client
	}
)

func (c client) Execute(ctx context.Context, method, url string, body, result any, opts ...Option) error {
	return c.execute(ctx, method, url, body, result, opts...)
}

func (c client) Get(ctx context.Context, url string, body, result any, opts ...Option) error {
	return c.execute(ctx, http.MethodGet, url, body, result, opts...)
}

func (c client) Post(ctx context.Context, url string, body, result any, opts ...Option) error {
	return c.execute(ctx, http.MethodPost, url, body, result, opts...)
}

func (c client) Delete(ctx context.Context, url string, body, result any, opts ...Option) error {
	return c.execute(ctx, http.MethodDelete, url, body, result, opts...)
}

func (c client) Patch(ctx context.Context, url string, body, result any, opts ...Option) error {
	return c.execute(ctx, http.MethodPatch, url, body, result, opts...)
}

func (c client) applyOptions(request *resty.Request, opts []Option) {
	var options options
	for _, opt := range opts {
		opt.apply(&options)
	}
	if headers := options.headers; headers != nil {
		request.SetHeaders(headers)
	}
	if pathParams := options.pathParams; pathParams != nil {
		request.SetPathParams(pathParams)
	}
	request.SetDebug(options.debug)

}

func (c client) execute(ctx context.Context, method string, url string, body, result any, opts ...Option) error {
	request := c.restyClient.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(result)
	c.applyOptions(request, opts)
	_, err := request.Execute(method, url)
	return err
}

func NewClient(logger logger.Logger, host string, timeout time.Duration,
	customMiddlewares ...middleware.Middleware) Client {
	return NewGenericClient(logger, host, nil, timeout, customMiddlewares...)
}

func NewGenericClient(logger logger.Logger, host string, baseHeaders map[string]string, timeout time.Duration,
	customMiddlewares ...middleware.Middleware) Client {
	restyClient := resty.New()
	restyClient.SetBaseURL(host)
	if timeout > 0 {
		restyClient.SetTimeout(timeout)
	}
	if len(baseHeaders) > 0 {
		restyClient.SetHeaders(baseHeaders)
	}
	append(middleware.MiddleWares(logger), customMiddlewares...).Apply(restyClient)
	return &client{
		restyClient: restyClient,
	}
}
