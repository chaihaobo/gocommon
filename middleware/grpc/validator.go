package grpc

import (
	validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
)

func ValidatorUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return validator.UnaryServerInterceptor()
}

func ValidatorStreamServerInterceptor() grpc.StreamServerInterceptor {
	return validator.StreamServerInterceptor()
}

// WithValidation returns gRPC server options with request validator
func WithValidation() []grpc.ServerOption {
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(ValidatorUnaryServerInterceptor()),
		grpc.StreamInterceptor(ValidatorStreamServerInterceptor()),
	}
	return serverOptions
}
