package queuex

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"
	"google.golang.org/protobuf/proto"
)

// RedisQueue implements Queue interface using Redis.
type RedisQueue struct {
	client    *redis.Client
	queueName string
}

// NewRedisQueue creates a new RedisQueue instance.
func NewRedisQueue(client *redis.Client, queueName string) RedisQueue {
	return RedisQueue{client: client, queueName: queueName}
}

// Push adds a message to the queue.
func (r RedisQueue) Push(_ context.Context, data []byte) error {
	return r.client.LPush(r.queueName, data).Err()
}

// Pop retrieves a message from the queue.
func (r RedisQueue) Pop(_ context.Context) ([]byte, error) {
	res, err := r.client.BLPop(1*time.Second, r.queueName).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if len(res) != 2 {
		return nil, nil
	}
	return []byte(res[1]), nil
}

// JSONRedisHandler handles JSON messages in a Redis queue.
type JSONRedisHandler[T any] struct {
	handlerFn func(ctx context.Context, msg T) error
	JSONCodec[T]
	RedisQueue
}

// NewJSONRedisHandler creates a new JSONRedisHandler instance.
func NewJSONRedisHandler[T any](
	client *redis.Client,
	queueName string,
	handlerFn func(ctx context.Context, msg T) error,
) *JSONRedisHandler[T] {
	return &JSONRedisHandler[T]{
		JSONCodec:  JSONCodec[T]{},
		RedisQueue: NewRedisQueue(client, queueName),
		handlerFn:  handlerFn,
	}
}

// Handle processes a message from the queue.
func (h *JSONRedisHandler[T]) Handle(ctx context.Context, msg T) error {
	if h.handlerFn == nil {
		return nil
	}
	return h.handlerFn(ctx, msg)
}

// ProtoRedisHandler handles protobuf messages in a Redis queue.
type ProtoRedisHandler[T proto.Message] struct {
	handlerFn func(ctx context.Context, msg T) error
	ProtoCodec[T]
	RedisQueue
}

// NewProtoRedisHandler creates a new ProtoRedisHandler instance.
func NewProtoRedisHandler[T proto.Message](
	client *redis.Client,
	queueName string,
	handlerFn func(ctx context.Context, msg T) error,
	newMsgFn func() T,
) *ProtoRedisHandler[T] {
	return &ProtoRedisHandler[T]{
		ProtoCodec: ProtoCodec[T]{newMsg: newMsgFn},
		RedisQueue: NewRedisQueue(client, queueName),
		handlerFn:  handlerFn,
	}
}

// Handle processes a message from the queue.
func (h *ProtoRedisHandler[T]) Handle(ctx context.Context, msg T) error {
	if h.handlerFn == nil {
		return nil
	}
	return h.handlerFn(ctx, msg)
}
