package middleware

import (
	"net/http"

	"github.com/chaihaobo/gocommon/trace"
	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
)

type TraceMiddleware struct {
}

func (l *TraceMiddleware) PreRequestHook(client *resty.Client, request *http.Request) error {
	otelhttptrace.Inject(request.Context(), request, otelhttptrace.WithPropagators(trace.Propagator))
	return nil
}

func (l *TraceMiddleware) OnAfterResponse(client *resty.Client, response *resty.Response) error {
	return nil
}
func (l *TraceMiddleware) OnError(client *resty.Request, err error) {
	// not on error
}

func NewTraceMiddleware() Middleware {
	return &TraceMiddleware{}
}
