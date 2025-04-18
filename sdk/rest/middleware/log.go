package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/chaihaobo/gocommon/logger"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type LoggingMiddleware struct {
	logger logger.Logger
}

func (l *LoggingMiddleware) PreRequestHook(client *resty.Client, request *http.Request) error {
	bodyBytes := make([]byte, 0)
	if request.Body != nil {
		bodyBytes, _ = io.ReadAll(request.Body)
	}
	l.logger.Info(request.Context(), "request",
		zap.String("host", client.BaseURL),
		zap.String("method", request.Method),
		zap.String("url", request.URL.RequestURI()),
		zap.Any("headers", request.Header),
		zap.String("body", string(bodyBytes)),
	)
	request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return nil
}

func (l *LoggingMiddleware) OnAfterResponse(client *resty.Client, response *resty.Response) error {
	l.logger.Info(response.Request.Context(), "response",
		zap.String("status", response.Status()),
		zap.Any("body", string(response.Body())),
		zap.String("timeused", response.Time().String()),
	)
	return nil
}
func (l *LoggingMiddleware) OnError(client *resty.Request, err error) {
	// not on error
}

func NewLoggingMiddleware(logger logger.Logger) Middleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}
