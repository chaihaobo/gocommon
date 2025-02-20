package restkit

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"

	"github.com/chaihaobo/gocommon/constant"
	commonErr "github.com/chaihaobo/gocommon/error"
)

var (
	DefaultTranslator = ut.New(en.New()).GetFallback()
)

type (
	responseBody struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	}
)

func newResponseBody(code string, message string, data any) responseBody {
	return responseBody{code, message, data}
}

func normalResponseBody(data any) responseBody {
	return newResponseBody(constant.Successful.Code, constant.Successful.Message, data)
}

// HTTPWrite write a normal json response
func HTTPWrite(writer http.ResponseWriter, data any) {
	writer.WriteHeader(http.StatusOK)
	appendGenericHeader(writer)
	writer.Write(lo.Must(json.Marshal(normalResponseBody(data))))
}

// HTTPWriteErr write an error json body
func HTTPWriteErr(writer http.ResponseWriter, err error) {
	var (
		serviceErr = constant.ErrSystemMalfunction
	)
	switch err := err.(type) {
	case commonErr.ServiceError:
		serviceErr = err
	case validator.ValidationErrors:
		serviceErr = commonErr.ServiceError{
			Code: constant.ErrorBadRequest.Code,
			Message: lo.Entries[string, string](err.
				Translate(DefaultTranslator))[0].Value,
		}
	default:
		serviceErr = constant.ErrSystemMalfunction
	}
	_ = errors.As(err, &serviceErr)
	actualHTTPStatus, ok := constant.ServiceErrorCode2HTTPStatus[serviceErr.Code]
	if !ok {
		actualHTTPStatus = http.StatusOK
	}

	for key, value := range serviceErr.Attributes {
		writer.Header().Add(key, value)
	}
	appendGenericHeader(writer)
	writer.WriteHeader(actualHTTPStatus)
	writer.Write(lo.Must(json.Marshal(newResponseBody(serviceErr.Code, serviceErr.Message, nil))))
}

func appendGenericHeader(writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json;charset=utf-8")
}
