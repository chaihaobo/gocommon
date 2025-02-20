package rest

import (
	"sync"
	"time"

	"github.com/chaihaobo/gocommon/logger"
	"github.com/chaihaobo/gocommon/sdk/rest/middleware"
)

type (
	ClientFactory interface {
		LoadClient(host string, timeout time.Duration, customMiddlewares ...middleware.Middleware) Client
	}
	clientKey struct {
		host    string
		timeout time.Duration
	}
	clientFactory struct {
		logger  logger.Logger
		clients sync.Map
		mutex   sync.Mutex
	}
)

func (f *clientFactory) LoadClient(host string, timeout time.Duration, customMiddlewares ...middleware.Middleware) Client {
	if len(customMiddlewares) > 0 {
		return NewClient(f.logger, host, timeout, customMiddlewares...)
	}

	key := newClientKey(host, timeout)
	if value, ok := f.clients.Load(key); ok {
		return value.(Client)
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if value, ok := f.clients.Load(key); ok {
		return value.(Client)
	}
	client := NewClient(f.logger, host, timeout)
	f.clients.Store(key, client)
	return client
}

func NewClientFactory(logger logger.Logger) ClientFactory {
	return &clientFactory{
		logger: logger,
	}
}

func newClientKey(host string, timeout time.Duration) clientKey {
	return clientKey{
		host:    host,
		timeout: timeout,
	}
}
