package grpc

import (
	validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// WithDefault returns default gRPC server option with validation and recovery
func WithDefault(errorMapper map[string]codes.Code) []grpc.ServerOption {
	unaryRecovery, streamRecovery := recoveryInterceptor()
	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			validator.UnaryServerInterceptor(),
			unaryRecovery,
			ErrorMappingUnaryServerInterceptor(errorMapper),
		),
		grpc.ChainStreamInterceptor(
			validator.StreamServerInterceptor(),
			streamRecovery,
			ErrorMappingStreamServerInterceptor(errorMapper),
		)}
	return serverOptions
}
