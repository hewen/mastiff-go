// Package server provides a simple queue server implementation
package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"github.com/panjf2000/ants/v2"
	"google.golang.org/protobuf/proto"
)

const (
	// DefaultQueueGoroutinePoolSize is the default size of the goroutine pool for processing queue messages.
	DefaultQueueGoroutinePoolSize = 1000
)

// QueueMessage is an interface for queue messages.
type QueueMessage any

// QueueServer is a simple queue server that processes messages from a queue using a goroutine pool.
type QueueServer[T any] struct {
	done               chan struct{}
	handler            QueueHandler[T]
	pool               *ants.Pool
	logger             logger.Logger
	poolSize           int
	EmptySleepInterval time.Duration
}

// Codec interface: encode/decode message.
type Codec[T any] interface {
	Encode(msg T) ([]byte, error)
	Decode(data []byte) (T, error)
}

// Queue interface: push/pop raw []byte data.
type Queue interface {
	Push(ctx context.Context, data []byte) error
	Pop(ctx context.Context) ([]byte, error)
}

// Handler interface: handle decoded message.
type Handler[T any] interface {
	Handle(ctx context.Context, msg T) error
}

// QueueHandler defines the interface for handling queue messages.
type QueueHandler[T any] interface {
	Codec[T]
	Queue
	Handler[T]
}

// JSONCodec implements Codec interface with JSON.
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

// ProtoCodec implements Codec interface with protobuf.
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
func NewQueueServer[T any](conf QueueConf, handler QueueHandler[T]) (*QueueServer[T], error) {
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
	qs.logger.Infof("Start queue service")

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
	qs.logger.Infof("Shutdown queue service")
	select {
	case <-qs.done:
		return
	default:
		close(qs.done)
	}

	for qs.pool.Running() == 0 {
		break
	}

	qs.pool.Release()
}

// runOnce retrieves a message from the queue and processes it using the goroutine pool.
func (qs *QueueServer[T]) runOnce(ctx context.Context) error {
	data, err := qs.handler.Pop(ctx)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		time.Sleep(qs.EmptySleepInterval)
		return nil
	}

	msg, err := qs.handler.Decode(data)
	if err != nil {
		return err
	}

	err = qs.pool.Submit(func() {
		if err := qs.handler.Handle(ctx, msg); err != nil {
			qs.logger.Errorf("failed to handle message: %v", err)
		}
		qs.logger.Infof("push success! => goroutine pool: [cap: %d, running: %d, free: %d]", qs.pool.Cap(), qs.pool.Running(), qs.pool.Free())
	})
	if err != nil {
		qs.logger.Errorf("submit to goroutine pool failed: %v", err)
	}

	return nil
}
