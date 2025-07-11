package server

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MyTestMsg struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

type mockQueueHandler struct {
	JSONCodec[MyTestMsg]

	handleFn    func(context.Context, MyTestMsg) error
	messages    [][]byte
	handled     []MyTestMsg
	handleDelay time.Duration
	popIndex    int
	mu          sync.Mutex
}

func (h *mockQueueHandler) Push(ctx context.Context, data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messages = append(h.messages, data)
	logger.NewLoggerWithContext(ctx).Infof("push => %v", string(data))
	return nil
}

func (h *mockQueueHandler) Pop(ctx context.Context) ([]byte, error) {
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

func (h *mockQueueHandler) Handle(ctx context.Context, msg MyTestMsg) error {
	if h.handleDelay > 0 {
		time.Sleep(h.handleDelay)
	}
	h.mu.Lock()
	h.handled = append(h.handled, msg)
	h.mu.Unlock()

	if h.handleFn != nil {
		return h.handleFn(ctx, msg)
	}
	return nil
}

func TestQueueServer_Messages(t *testing.T) {
	handler := &mockQueueHandler{}

	conf := QueueConf{
		PoolSize:           5,
		EmptySleepInterval: 1 * time.Millisecond,
	}
	server, err := NewQueueServer(conf, handler)
	require.NoError(t, err)

	ctx := context.Background()
	msgs := []MyTestMsg{
		{ID: 1, Body: "hello"},
		{ID: 2, Body: "world"},
	}
	for _, m := range msgs {
		data, err := handler.Encode(m)
		require.NoError(t, err)
		err = handler.Push(ctx, data)
		require.NoError(t, err)
	}

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

func TestQueueServer_BulkMessages(t *testing.T) {
	const total = 100
	handler := &mockQueueHandler{}

	conf := QueueConf{
		PoolSize:           10,
		EmptySleepInterval: 1 * time.Millisecond,
	}
	server, err := NewQueueServer(conf, handler)
	require.NoError(t, err)

	ctx := context.Background()
	for i := 0; i < total; i++ {
		msg := MyTestMsg{ID: i, Body: "msg"}
		data, err := handler.Encode(msg)
		require.NoError(t, err)
		err = handler.Push(ctx, data)
		require.NoError(t, err)
	}

	go server.Start()
	defer server.Stop()

	time.Sleep(200 * time.Millisecond)

	handler.mu.Lock()
	defer handler.mu.Unlock()
	assert.Len(t, handler.handled, total)
}

type popErrorHandler struct {
	JSONCodec[MyTestMsg]
}

func (h *popErrorHandler) Push(_ context.Context, _ []byte) error { return nil }
func (h *popErrorHandler) Pop(_ context.Context) ([]byte, error)  { return nil, errors.New("pop error") }
func (h *popErrorHandler) Handle(_ context.Context, _ MyTestMsg) error {
	return nil
}

func TestQueueServer_PopError(t *testing.T) {
	handler := &popErrorHandler{}
	conf := QueueConf{PoolSize: 1, EmptySleepInterval: 10 * time.Millisecond}
	server, err := NewQueueServer(conf, handler)
	assert.Nil(t, err)

	go server.Start()
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)
}

type decodeErrorHandler struct {
	JSONCodec[MyTestMsg]
	messages [][]byte
}

func (h *decodeErrorHandler) Push(_ context.Context, data []byte) error {
	h.messages = append(h.messages, data)
	return nil
}
func (h *decodeErrorHandler) Pop(_ context.Context) ([]byte, error) {
	if len(h.messages) == 0 {
		return nil, nil
	}
	data := h.messages[0]
	h.messages = h.messages[1:]
	return data, nil
}
func (h *decodeErrorHandler) Decode(_ []byte) (MyTestMsg, error) {
	return MyTestMsg{}, errors.New("decode error")
}
func (h *decodeErrorHandler) Handle(_ context.Context, _ MyTestMsg) error {
	return nil
}

func TestQueueServer_DecodeError(t *testing.T) {
	handler := &decodeErrorHandler{}
	_ = handler.Push(context.Background(), []byte("invalid json"))

	conf := QueueConf{PoolSize: 1, EmptySleepInterval: 10 * time.Millisecond}
	server, err := NewQueueServer(conf, handler)
	assert.Nil(t, err)

	go server.Start()
	defer server.Stop()
	time.Sleep(50 * time.Millisecond)
}

type handleErrorHandler struct {
	JSONCodec[MyTestMsg]
	messages [][]byte
}

func (h *handleErrorHandler) Push(_ context.Context, data []byte) error {
	h.messages = append(h.messages, data)
	return nil
}
func (h *handleErrorHandler) Pop(_ context.Context) ([]byte, error) {
	if len(h.messages) == 0 {
		return nil, nil
	}
	data := h.messages[0]
	h.messages = h.messages[1:]
	return data, nil
}
func (h *handleErrorHandler) Handle(_ context.Context, _ MyTestMsg) error {
	return errors.New("handle error")
}

func TestQueueServer_HandleError(t *testing.T) {
	handler := &handleErrorHandler{}
	data, _ := handler.Encode(MyTestMsg{ID: 1, Body: "fail"})
	err := handler.Push(context.Background(), data)
	assert.Nil(t, err)

	conf := QueueConf{PoolSize: 1, EmptySleepInterval: 10 * time.Millisecond}
	server, err := NewQueueServer(conf, handler)
	assert.Nil(t, err)

	go server.Start()
	defer server.Stop()
	time.Sleep(50 * time.Millisecond)
}

func TestQueueServer_StartStopExit(t *testing.T) {
	handler := &mockQueueHandler{}
	conf := QueueConf{
		PoolSize:           2,
		EmptySleepInterval: 1 * time.Millisecond,
	}
	server, err := NewQueueServer(conf, handler)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		server.Start()
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	server.Stop()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Start did not exit after Stop called")
	}
}

func TestQueueServer_StopIdempotent(t *testing.T) {
	handler := &mockQueueHandler{}
	server, err := NewQueueServer(QueueConf{}, handler)
	require.NoError(t, err)

	server.Stop()
	server.Stop() // safe second call
}
