package queue

import (
	"context"
	"encoding"
	"fmt"
	"github.com/chaihaobo/gocommon/logger"
	"reflect"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type RedisQueue struct {
	logger        logger.Logger
	asynqServer   *asynq.Server
	asynqClient   *asynq.Client
	asynqServeMux *asynq.ServeMux
}

func (r *RedisQueue) Publish(ctx context.Context, topic string, message any, opts ...Option) error {
	data, err := r.parseMessageToBytes(message)
	if err != nil {
		return err
	}
	asynqOptions := r.mappingAsynqOptions(opts)
	taskInfo, err := r.asynqClient.EnqueueContext(ctx, asynq.NewTask(topic, data), asynqOptions...)
	if err != nil {
		return err
	}
	r.logger.Info(ctx, "published message to redis queue success",
		zap.ByteString("payload", data), zap.String("id", taskInfo.ID))
	return nil
}

func (r *RedisQueue) SubscribeTo(topic string, subscriber Subscriber) {
	r.asynqServeMux.HandleFunc(topic, func(ctx context.Context, task *asynq.Task) error {
		topic := task.Type()
		payload := task.Payload()
		return subscriber.Subscribe(ctx, topic, payload)
	})
}

func (r *RedisQueue) parseMessageFromBytes(subscriberFunc SubscriberFunc, payload []byte) (any, error) {
	var result any
	subscribeFunc := reflect.TypeOf(subscriberFunc)
	messageType := subscribeFunc.In(2)

	if messageType.Implements(reflect.TypeOf(new(encoding.BinaryUnmarshaler)).Elem()) {
		if messageType.Kind() == reflect.Ptr {
			messageType = messageType.Elem()
		}
		result = (reflect.New(messageType).Interface().(encoding.BinaryUnmarshaler)).UnmarshalBinary(payload)
		return result, nil
	}
	if messageType.Kind() == reflect.Slice && messageType.Elem().Kind() == reflect.Uint8 {
		result = payload
		return result, nil
	}
	return nil, fmt.Errorf("unsupported message type: %T, only support []byte|encoding.BinaryUnmarshaler", messageType)
}

func (r *RedisQueue) StartSubscriber() error {
	return r.asynqServer.Start(r.asynqServeMux)
}

func (r *RedisQueue) RunSubscriber() error {
	return r.asynqServer.Run(r.asynqServeMux)
}

func (r *RedisQueue) Shutdown() {
	r.asynqServer.Shutdown()
}

func (r *RedisQueue) mappingAsynqOptions(opts []Option) []asynq.Option {
	options := &options{}
	for _, opt := range opts {
		opt.apply(options)
	}
	asynqOpts := make([]asynq.Option, 0)
	if options.delayDuration > 0 {
		asynqOpts = append(asynqOpts, asynq.ProcessAt(time.Now().Add(options.delayDuration)))
	}
	return asynqOpts
}

func (r *RedisQueue) parseMessageToBytes(message any) ([]byte, error) {
	switch result := message.(type) {
	case []byte:
		return result, nil
	case encoding.BinaryMarshaler:
		return result.MarshalBinary()
	default:
		return nil, fmt.Errorf("unsupported message type: %T, only support []byte|encoding.BinaryMarshaler", message)
	}

}

func NewRedisQueue(logger logger.Logger, address string, db int, password string) (Queue, error) {
	redisClientOpt := asynq.RedisClientOpt{
		Addr:     address,
		DB:       db,
		Password: password,
	}
	server := asynq.NewServer(redisClientOpt, asynq.Config{})
	client := asynq.NewClient(redisClientOpt)
	return &RedisQueue{
		logger:        logger,
		asynqServer:   server,
		asynqClient:   client,
		asynqServeMux: asynq.NewServeMux(),
	}, nil
}
