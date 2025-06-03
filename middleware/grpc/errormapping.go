package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	commonErr "github.com/chaihaobo/gocommon/error"
)

const (
	errorCode    = "error_code"
	errorMessage = "error_message"
)

// ErrorMappingUnaryServerInterceptor returns a new unary server interceptor that added error detail for service error.
func ErrorMappingUnaryServerInterceptor(errorMapper map[string]codes.Code) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			switch serviceError := err.(type) {
			case commonErr.ServiceError:
				md := errorMetadata(serviceError)
				if err := grpc.SetTrailer(ctx, md); err != nil {
					slog.ErrorContext(ctx, "failed to set trailer", slog.Any("error", err))
				}
				statusCode, found := errorMapper[serviceError.Code]
				if !found {
					statusCode = codes.Unknown
				}
				return nil, grpcError(statusCode, serviceError)
			default:
				return nil, err
			}
		}

		return resp, nil
	}
}

// ErrorMappingStreamServerInterceptor returns a new streaming server interceptor that added error detail for service error.
func ErrorMappingStreamServerInterceptor(errorMapper map[string]codes.Code) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream,
		info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		if err != nil {
			switch serviceError := err.(type) {
			case commonErr.ServiceError:
				md := errorMetadata(serviceError)
				stream.SetTrailer(md)
				statusCode, found := errorMapper[serviceError.Code]
				if !found {
					statusCode = codes.Unknown
				}
				return grpcError(statusCode, serviceError)
			default:
				return err
			}
		}
		return nil
	}
}

func grpcError(statusCode codes.Code, serviceError commonErr.ServiceError) error {
	pbError := commonErr.Error{
		Code:       serviceError.Code,
		Message:    serviceError.Message,
		Attributes: serviceError.Attributes,
	}
	grpcError := status.New(statusCode, serviceError.Message)
	errWithDetails, err := grpcError.WithDetails(&pbError)
	if err != nil {
		return grpcError.Err()
	}
	return errWithDetails.Err()
}

func errorMetadata(serviceError commonErr.ServiceError) metadata.MD {
	data := make(map[string]string)
	data[errorCode] = serviceError.Code
	data[errorMessage] = serviceError.Message
	if len(serviceError.Attributes) > 0 {
		for k, v := range serviceError.Attributes {
			data[k] = v
		}
	}
	return metadata.New(data)
}

// WithErrorMapping returns gRPC server options with request validator
func WithErrorMapping(errorMapper map[string]codes.Code) []grpc.ServerOption {
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(ErrorMappingUnaryServerInterceptor(errorMapper)),
		grpc.StreamInterceptor(ErrorMappingStreamServerInterceptor(errorMapper)),
	}
	return serverOptions
}
