package error

import "errors"

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
