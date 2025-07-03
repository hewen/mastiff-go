package server

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"github.com/stretchr/testify/assert"
)

type MyTestMsg struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type mockHandler struct {
	messages    [][]byte
	handled     []MyTestMsg
	mu          sync.Mutex
	popIndex    int
	handleDelay time.Duration
}

func (h *mockHandler) Encode(msg MyTestMsg) ([]byte, error) {
	return json.Marshal(msg)
}

func (h *mockHandler) Decode(data []byte) (MyTestMsg, error) {
	var m MyTestMsg
	err := json.Unmarshal(data, &m)
	return m, err
}

func (h *mockHandler) Push(ctx context.Context, msg MyTestMsg) error {
	logger.NewLoggerWithContext(ctx).Infof("push => %#v", msg)
	data, err := h.Encode(msg)
	if err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messages = append(h.messages, data)
	return nil
}

func (h *mockHandler) Pop(ctx context.Context) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.popIndex >= len(h.messages) {
		return nil, nil
	}
	data := h.messages[h.popIndex]
	h.popIndex++
	logger.NewLoggerWithContext(ctx).Infof("pop <= %v", string(data))
	return data, nil
}

func (h *mockHandler) Handle(_ context.Context, msg MyTestMsg) error {
	if h.handleDelay > 0 {
		time.Sleep(h.handleDelay)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handled = append(h.handled, msg)
	return nil
}

func TestQueueServer_Messages(t *testing.T) {
	handler := &mockHandler{}
	server, err := NewQueueServer(handler, 0)
	if err != nil {
		t.Fatalf("failed to create queue server: %v", err)
	}

	ctx := context.Background()
	err = handler.Push(ctx, MyTestMsg{ID: 1, Body: "hello"})
	assert.Nil(t, err)
	err = handler.Push(ctx, MyTestMsg{ID: 2, Body: "world"})
	assert.Nil(t, err)

	go server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	handler.mu.Lock()
	defer handler.mu.Unlock()

	var gotBodies []string
	for _, msg := range handler.handled {
		gotBodies = append(gotBodies, msg.Body)
	}
	assert.ElementsMatch(t, gotBodies, []string{"hello", "world"})
}

type wrapHandler[T any] struct {
	inner QueueHandler[T]
	wg    *sync.WaitGroup
}

func (w *wrapHandler[T]) Encode(msg T) ([]byte, error) {
	return w.inner.Encode(msg)
}

func (w *wrapHandler[T]) Decode(data []byte) (T, error) {
	return w.inner.Decode(data)
}

func (w *wrapHandler[T]) Push(ctx context.Context, msg T) error {
	return w.inner.Push(ctx, msg)
}

func (w *wrapHandler[T]) Pop(ctx context.Context) ([]byte, error) {
	return w.inner.Pop(ctx)
}

func (w *wrapHandler[T]) Handle(ctx context.Context, msg T) error {
	defer w.wg.Done()
	logger.NewLoggerWithContext(ctx).Infof("handle message: %+v", msg) // 确认调用
	return w.inner.Handle(ctx, msg)
}

func TestQueueServer_BulkMessages(t *testing.T) {
	const totalMsgs = 200

	origHandler := &mockHandler{}
	poolSize := 4

	server, err := NewQueueServer[MyTestMsg](nil, poolSize)
	if err != nil {
		t.Fatalf("failed to create queue server: %v", err)
	}

	ctx := context.Background()

	for i := 0; i < totalMsgs; i++ {
		err := origHandler.Push(ctx, MyTestMsg{ID: i, Body: "msg"})
		if err != nil {
			t.Fatalf("push failed: %v", err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(totalMsgs)

	wrappedHandler := &wrapHandler[MyTestMsg]{
		inner: origHandler,
		wg:    &wg,
	}

	server.handler = wrappedHandler

	go server.Start()
	defer server.Stop()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for all messages to be processed")
	}

	origHandler.mu.Lock()
	defer origHandler.mu.Unlock()

	if len(origHandler.handled) != totalMsgs {
		t.Errorf("expected %d handled messages, got %d", totalMsgs, len(origHandler.handled))
	}
}

func TestNewQueueServer_PoolError(t *testing.T) {
	_, err := NewQueueServer(&mockHandler{}, -1)
	assert.NoError(t, err)
}

type popErrorHandler struct {
	QueueHandler[MyTestMsg]
}

func (h *popErrorHandler) Pop(_ context.Context) ([]byte, error) {
	return nil, errors.New("pop error")
}

func TestQueueServer_PopError(_ *testing.T) {
	handler := &popErrorHandler{}
	server, _ := NewQueueServer(handler, 1)

	go server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)
}

type decodeErrorHandler struct {
	QueueHandler[MyTestMsg]
}

func (h *decodeErrorHandler) Pop(_ context.Context) ([]byte, error) {
	return []byte("invalid json"), nil
}

func (h *decodeErrorHandler) Decode(_ []byte) (MyTestMsg, error) {
	return MyTestMsg{}, errors.New("decode error")
}

func TestQueueServer_DecodeError(_ *testing.T) {
	handler := &decodeErrorHandler{}
	server, _ := NewQueueServer(handler, 1)

	go server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)
}

type handleErrorHandler struct {
	QueueHandler[MyTestMsg]
}

func (h *handleErrorHandler) Pop(_ context.Context) ([]byte, error) {
	data, _ := json.Marshal(MyTestMsg{ID: 1, Body: "test"})
	return data, nil
}

func (h *handleErrorHandler) Decode(data []byte) (MyTestMsg, error) {
	var msg MyTestMsg
	err := json.Unmarshal(data, &msg)
	return msg, err
}

func (h *handleErrorHandler) Handle(_ context.Context, _ MyTestMsg) error {
	return errors.New("handle error")
}

func TestQueueServer_HandleError(_ *testing.T) {
	handler := &handleErrorHandler{}
	server, _ := NewQueueServer(handler, 1)

	go server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)
}

func TestQueueServer_StartStopExit(t *testing.T) {
	handler := &mockHandler{}
	server, err := NewQueueServer(handler, 1)
	if err != nil {
		t.Fatalf("failed to create queue server: %v", err)
	}

	done := make(chan struct{})
	go func() {
		server.Start()
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	server.Stop()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Start did not exit after Stop called")
	}
}

func TestQueueServer_StopIdempotent(_ *testing.T) {
	handler := &mockHandler{}
	server, _ := NewQueueServer(handler, 1)

	server.Stop()
	server.Stop()
}
