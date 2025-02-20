package restkit

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/gin-gonic/gin"

	"github.com/chaihaobo/gocommon/constant"
)

func TestHandler(t *testing.T) {
	router := gin.New()
	router.POST("foo", AdaptToGinHandler(HandlerFunc[string](func(ctx *gin.Context) (string, error) {
		return "bar", nil
	})))
	testcases := []struct {
		name     string
		handler  HandlerFunc[any]
		want     string
		wantCode int
	}{
		{
			name: "when everything is ok",
			handler: HandlerFunc[any](func(ctx *gin.Context) (any, error) {
				return "ok", nil
			}),
			wantCode: http.StatusOK,
			want:     "{\"code\":\"0000000\",\"message\":\"successful\",\"data\":\"ok\"}",
		},
		{
			name: "when happen internal server error",
			handler: HandlerFunc[any](func(ctx *gin.Context) (any, error) {
				return nil, constant.ErrSystemMalfunction
			}),
			want:     "{\"code\":\"9999999\",\"message\":\"system malfunction\",\"data\":null}",
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		router := gin.New()
		router.POST("foo", AdaptToGinHandler(testcase.handler))
		request, _ := http.NewRequest(http.MethodPost, "/foo", bytes.NewReader([]byte("")))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		assert.Equal(t, testcase.wantCode, response.Code)
		assert.Equal(t, testcase.want, response.Body.String())

	}
}
