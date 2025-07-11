package server

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

// QueueJSONRedisHandler handles JSON messages in a Redis queue.
type QueueJSONRedisHandler[T any] struct {
	handlerFn func(ctx context.Context, msg T) error
	JSONCodec[T]
	RedisQueue
}

// NewQueueJSONRedisHandler creates a new QueueJSONRedisHandler instance.
func NewQueueJSONRedisHandler[T any](
	client *redis.Client,
	queueName string,
	handlerFn func(ctx context.Context, msg T) error,
) *QueueJSONRedisHandler[T] {
	return &QueueJSONRedisHandler[T]{
		JSONCodec:  JSONCodec[T]{},
		RedisQueue: NewRedisQueue(client, queueName),
		handlerFn:  handlerFn,
	}
}

// Handle processes a message from the queue.
func (h *QueueJSONRedisHandler[T]) Handle(ctx context.Context, msg T) error {
	if h.handlerFn == nil {
		return nil
	}
	return h.handlerFn(ctx, msg)
}

// QueueProtoRedisHandler handles protobuf messages in a Redis queue.
type QueueProtoRedisHandler[T proto.Message] struct {
	handlerFn func(ctx context.Context, msg T) error
	ProtoCodec[T]
	RedisQueue
}

// NewQueueProtoRedisHandler creates a new QueueProtoRedisHandler instance.
func NewQueueProtoRedisHandler[T proto.Message](
	client *redis.Client,
	queueName string,
	handlerFn func(ctx context.Context, msg T) error,
	newMsgFn func() T,
) *QueueProtoRedisHandler[T] {
	return &QueueProtoRedisHandler[T]{
		ProtoCodec: ProtoCodec[T]{newMsg: newMsgFn},
		RedisQueue: NewRedisQueue(client, queueName),
		handlerFn:  handlerFn,
	}
}

// Handle processes a message from the queue.
func (h *QueueProtoRedisHandler[T]) Handle(ctx context.Context, msg T) error {
	if h.handlerFn == nil {
		return nil
	}
	return h.handlerFn(ctx, msg)
}
