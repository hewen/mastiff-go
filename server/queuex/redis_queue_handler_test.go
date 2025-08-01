// Package queuex provides a queue server.
package queuex

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/server/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRedisHandler_Basic(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	var handledMsg *test.TestMsg

	handlerFn := func(_ context.Context, msg *test.TestMsg) error {
		handledMsg = msg
		return nil
	}
	newMsgFn := func() *test.TestMsg {
		return &test.TestMsg{}
	}

	qh := NewProtoRedisHandler(client, "myqueue", handlerFn, newMsgFn)

	ctx := context.Background()

	msg := &test.TestMsg{Id: 1, Name: "Alice"}
	pushData, err := qh.Encode(msg)
	require.NoError(t, err)
	err = qh.Push(ctx, pushData)
	require.NoError(t, err)

	popData, err := qh.Pop(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, popData)

	decodedMsg, err := qh.Decode(popData)
	require.NoError(t, err)
	require.True(t, proto.Equal(msg, decodedMsg))

	err = qh.Handle(ctx, decodedMsg)
	require.NoError(t, err)

	require.NotNil(t, handledMsg)
	require.Equal(t, msg.Id, handledMsg.Id)
	require.Equal(t, msg.Name, handledMsg.Name)
}

func TestRedisHandlerWithJSON(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	queueName := "json_test_queue"

	type MyMsg struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}

	handleFn := func(_ context.Context, _ MyMsg) error {
		return nil
	}

	handler := NewJSONRedisHandler(redisClient, queueName, handleFn)

	err = handler.Handle(context.TODO(), MyMsg{Name: "test"})
	assert.Nil(t, err)
}

func TestRedisHandlerWithProto(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	queueName := "proto_test_queue"
	var mu sync.Mutex
	var handledMsg *test.TestMsg

	handleFn := func(_ context.Context, msg *test.TestMsg) error {
		mu.Lock()
		defer mu.Unlock()
		handledMsg = msg
		return nil
	}

	handler := NewProtoRedisHandler(redisClient, queueName, handleFn, func() *test.TestMsg {
		return &test.TestMsg{}
	})

	qs, err := NewQueueServer(serverconf.QueueConfig{
		QueueName:          queueName,
		PoolSize:           10,
		EmptySleepInterval: 5 * time.Millisecond,
	}, handler)
	require.NoError(t, err)

	ctx := context.Background()

	go qs.Start()
	defer qs.Stop()

	msg := &test.TestMsg{Id: 123, Name: "Alice"}
	pushData, err := handler.Encode(msg)
	require.NoError(t, err)
	err = handler.Push(ctx, pushData)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	require.NotNil(t, handledMsg)
	require.True(t, proto.Equal(msg, handledMsg))
}

func TestRedisQueue_Pop_LengthMismatch(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	queue := NewRedisQueue(client, "test_queue_len")

	err = client.LPush("test_queue_len", "only-one").Err()
	require.NoError(t, err)

	client.LPop("test_queue_len")

	data, err := queue.Pop(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, data)
}

func TestJSONRedisHandler_Handle_NilHandler(t *testing.T) {
	handler := &JSONRedisHandler[MyTestMsg]{}

	err := handler.Handle(context.Background(), MyTestMsg{ID: 1, Body: "test"})
	assert.NoError(t, err)
}

func TestProtoRedisHandler_Handle_NilHandler(t *testing.T) {
	dummy := NewProtoRedisHandler(nil, "test", nil, func() *test.TestMsg {
		return &test.TestMsg{}
	})

	err := dummy.Handle(context.Background(), &test.TestMsg{})
	assert.NoError(t, err)
}
