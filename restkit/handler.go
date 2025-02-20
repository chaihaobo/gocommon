package restkit

import "github.com/gin-gonic/gin"

type (
	// Handler is a wrapper around gin.HandlerFunc, which returns a response and an error
	Handler[Response any] interface {
		Invoke(ctx *gin.Context) (Response, error)
	}
	HandlerFunc[Response any] func(ctx *gin.Context) (Response, error)
)

func (f HandlerFunc[Response]) Invoke(ctx *gin.Context) (Response, error) {
	return f(ctx)
}

// AdaptToGinHandler adapts a Handler to a gin.HandlerFunc
func AdaptToGinHandler[Response any](handler Handler[Response]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		response, err := handler.Invoke(ctx)
		if err != nil {
			HTTPWriteErr(ctx.Writer, err)
			return
		}
		HTTPWrite(ctx.Writer, response)
	}
}
