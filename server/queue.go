// Package server provides a simple queue server implementation
package server

import (
	"context"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"github.com/panjf2000/ants/v2"
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
	logger             *logger.Logger
	poolSize           int
	EmptySleepInterval time.Duration
}

// QueueHandler defines the interface for handling queue messages.
type QueueHandler[T any] interface {
	Encode(msg T) ([]byte, error)
	Decode(data []byte) (T, error)
	Push(ctx context.Context, msg T) error
	Pop(ctx context.Context) ([]byte, error)
	Handle(ctx context.Context, msg T) error
}

// NewQueueServer creates a new QueueServer with the specified handler and pool size.
func NewQueueServer[T any](conf QueueConfig, handler QueueHandler[T]) (*QueueServer[T], error) {
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
