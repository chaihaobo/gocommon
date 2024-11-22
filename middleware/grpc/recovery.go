package grpc

import (
	"log"

	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func recoveryInterceptor() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	handler := func(p interface{}) (err error) {
		log.Print("panic", p)
		return status.Error(codes.Internal, "server panic")
	}

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(handler),
	}
	return recovery.UnaryServerInterceptor(opts...), recovery.StreamServerInterceptor(opts...)
}

func RecoveryUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	unaryServerInterceptor, _ := recoveryInterceptor()
	return unaryServerInterceptor
}

func RecoveryStreamServerInterceptor() grpc.StreamServerInterceptor {
	_, streamServerInterceptor := recoveryInterceptor()
	return streamServerInterceptor
}

// WithRecovery return gRPC server options with recovery handler
func WithRecovery() []grpc.ServerOption {
	unary, stream := recoveryInterceptor()
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(unary),
		grpc.StreamInterceptor(stream),
	}
	return serverOptions
}
