package error

import (
	"errors"
	"net/http"
)

type ServiceError struct {
	Code    string
	Message string
	//	Attributes will attach to grpc trailer metadata or http response header
	Attributes map[string]string
}

func (e ServiceError) Error() string {
	return e.Message
}

func (e ServiceError) Is(tgt error) bool {
	target := ServiceError{}
	ok := errors.As(tgt, &target)
	if !ok {
		return false
	}

	return e.Message == target.Message && e.Code == target.Code
}

func (s ServiceError) AttachToResponse(writer http.ResponseWriter) {
	writer.Header().Add("error_code", s.Code)
	writer.Header().Add("error_message", s.Message)
	for k, v := range s.Attributes {
		writer.Header().Add(k, v)
	}
}
