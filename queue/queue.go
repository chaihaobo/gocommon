package queue

import (
	"context"
	"encoding"
	"fmt"
	"github.com/chaihaobo/gocommon/trace"
	"go.opentelemetry.io/otel"
	"reflect"
)

type (
	Queue interface {
		// Publish 发布消息到topic中
		Publish(ctx context.Context, topic string, message any, opts ...Option) error
		// SubscribeTo 注册订阅者到主题中
		SubscribeTo(topic string, subscriber Subscriber)
		// StartSubscriber 异步启动订阅. 开始监听消息
		StartSubscriber() error
		// RunSubscriber 同步启动订阅. 开始监听消息
		RunSubscriber() error
		// Shutdown 停止订阅. 停止监听消息
		Shutdown()
	}
	Subscriber interface {
		Subscribe(ctx context.Context, topic string, message []byte) error
	}

	SubscriberFunc func(ctx context.Context, topic string, message []byte) error
)

func (s SubscriberFunc) Subscribe(ctx context.Context, topic string, message []byte) error {
	return s(ctx, topic, message)
}

// CreateSubscriber 通过处理函数返回订阅者接口
// T 必须为[]byte 或者 实现了encoding.BinaryUnmarshaler接口的指针类型
func CreateSubscriber[T any](handleFunc func(ctx context.Context, topic string, message T) error) Subscriber {
	return SubscriberFunc(func(ctx context.Context, topic string, payload []byte) error {
		ctx, span := otel.Tracer(trace.DefaultTracerName).Start(ctx, "queue.subscribe.consume."+topic)
		defer span.End()

		var handleFuncMsgArg T
		handleFuncType := reflect.TypeOf(handleFunc)
		messageType := handleFuncType.In(2)
		if messageType.Implements(reflect.TypeOf(new(encoding.BinaryUnmarshaler)).Elem()) &&
			messageType.Kind() == reflect.Ptr {
			result := reflect.New(messageType.Elem()).Interface().(encoding.BinaryUnmarshaler)
			if err := result.UnmarshalBinary(payload); err != nil {
				return err
			}
			handleFuncMsgArg = result.(T)
		}
		if messageType.Kind() == reflect.Slice && messageType.Elem().Kind() == reflect.Uint8 {
			handleFuncMsgArg = reflect.ValueOf(payload).Interface().(T)
		}
		if reflect.ValueOf(handleFuncMsgArg).IsZero() {
			return fmt.Errorf("message type not supported")
		}
		return handleFunc(ctx, topic, handleFuncMsgArg)
	})
}
