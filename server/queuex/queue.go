// Package queuex provides a queue server implementation.
package queuex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/panjf2000/ants/v2"
	"google.golang.org/protobuf/proto"
)

const (
	// DefaultQueueGoroutinePoolSize is the default size of the goroutine pool for processing queue messages.
	DefaultQueueGoroutinePoolSize = 1000
)

// ErrEmptyQueueName is returned when the queue name is empty.
var ErrEmptyQueueName = errors.New("queue name is empty")

// QueueMessage is an interface for queue messages.
type QueueMessage any

// QueueServer is a queue server that processes messages from a queue using a goroutine pool. It is used to provide a queue server implementation.
type QueueServer[T any] struct {
	handler            ServerHandler[T]
	logger             logger.Logger
	pool               *ants.Pool
	done               chan struct{}
	name               string
	EmptySleepInterval time.Duration
	poolSize           int
}

// Codec interface: encode/decode message. It is used to provide a codec for queue messages.
type Codec[T any] interface {
	Encode(msg T) ([]byte, error)
	Decode(data []byte) (T, error)
}

// Queue interface: push/pop raw []byte data. It is used to provide a queue for queue messages.
type Queue interface {
	Push(ctx context.Context, data []byte) error
	Pop(ctx context.Context) ([]byte, error)
}

// Handler interface: handle decoded message. It is used to provide a handler for queue messages.
type Handler[T any] interface {
	Handle(ctx context.Context, msg T) error
}

// ServerHandler defines the interface for handling queue messages.
type ServerHandler[T any] interface {
	Codec[T]
	Queue
	Handler[T]
}

// JSONCodec implements Codec interface with JSON. It is used to provide a codec for queue messages.
type JSONCodec[T any] struct{}

// Encode implements Codec interface.
func (c JSONCodec[T]) Encode(msg T) ([]byte, error) {
	return json.Marshal(msg)
}

// Decode implements Codec interface.
func (c JSONCodec[T]) Decode(data []byte) (T, error) {
	var msg T
	err := json.Unmarshal(data, &msg)
	return msg, err
}

// ProtoCodec implements Codec interface with protobuf. It is used to provide a codec for queue messages.
type ProtoCodec[T proto.Message] struct {
	newMsg func() T
}

// Encode implements Codec interface.
func (c ProtoCodec[T]) Encode(msg T) ([]byte, error) {
	return proto.Marshal(msg)
}

// Decode implements Codec interface.
func (c ProtoCodec[T]) Decode(data []byte) (T, error) {
	msg := c.newMsg()
	err := proto.Unmarshal(data, msg)
	return msg, err
}

// NewQueueServer creates a new QueueServer with the specified handler and pool size.
func NewQueueServer[T any](conf serverconf.QueueConfig, handler ServerHandler[T]) (*QueueServer[T], error) {
	if conf.QueueName == "" {
		return nil, ErrEmptyQueueName
	}

	if conf.PoolSize <= 0 {
		conf.PoolSize = DefaultQueueGoroutinePoolSize
	}
	if conf.EmptySleepInterval <= 0 {
		conf.EmptySleepInterval = 10 * time.Millisecond
	}

	log := logger.NewLogger()
	log.Infof("init goroutine pool size: %d", conf.PoolSize)

	pool, err := ants.NewPool(conf.PoolSize)
	if err != nil {
		return nil, err
	}

	return &QueueServer[T]{
		name:               conf.QueueName,
		done:               make(chan struct{}, 1),
		pool:               pool,
		handler:            handler,
		logger:             log,
		poolSize:           conf.PoolSize,
		EmptySleepInterval: conf.EmptySleepInterval,
	}, nil
}

// Start begins the queue server, processing messages at regular intervals.
func (qs *QueueServer[T]) Start() {
	ctx := context.Background()
	for {
		select {
		case <-qs.done:
			return
		default:
			if err := qs.runOnce(ctx); err != nil {
				qs.logger.Errorf("error: %v", err)
			}
		}
	}
}

// Stop stops the queue server and releases resources.
func (qs *QueueServer[T]) Stop() {
	select {
	case <-qs.done:
		return
	default:
		close(qs.done)
	}

	qs.ShutdownWait()
}

// ShutdownWait waits for all goroutines to finish and releases the pool.
func (qs *QueueServer[T]) ShutdownWait() {
	for qs.pool.Running() > 0 {
		time.Sleep(10 * time.Millisecond)
	}
	qs.pool.Release()
}

// Name returns the name of the queue server.
func (qs *QueueServer[T]) Name() string {
	return fmt.Sprintf("queue server (%s)", qs.name)
}

// WithLogger sets the logger for the queue server.
func (qs *QueueServer[T]) WithLogger(l logger.Logger) {
	if l != nil {
		qs.logger = l
	}
}

// runOnce retrieves a message from the queue and processes it using the goroutine pool.
func (qs *QueueServer[T]) runOnce(ctx context.Context) error {
	data, err := qs.handler.Pop(ctx)
	if err != nil {
		return fmt.Errorf("pop failed: %w", err)
	}
	if len(data) == 0 {
		time.Sleep(qs.EmptySleepInterval)
		return nil
	}

	msg, err := qs.handler.Decode(data)
	if err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	err = qs.pool.Submit(func() {
		if handleErr := qs.handler.Handle(ctx, msg); handleErr != nil {
			qs.logger.Errorf("[queue:%s] failed to handle message: %v", qs.name, handleErr)
		}
		qs.logger.Infof("submit to pool success => [queue: %s, cap: %d, running: %d, free: %d]", qs.name, qs.pool.Cap(), qs.pool.Running(), qs.pool.Free())
	})
	if err != nil {
		qs.logger.Errorf("submit to goroutine pool failed: %v", err)
	}

	return nil
}
