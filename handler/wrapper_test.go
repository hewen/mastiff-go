package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type FooRequest struct {
	Name string `json:"name"`
}

type FooResponse struct {
	Message string `json:"message"`
}

func FooHandler(_ Context, req FooRequest) (FooResponse, error) {
	return FooResponse{Message: "Hello, " + req.Name}, nil
}

func TestWrapHandlerGin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/foo", WrapHandlerGin(FooHandler))

	reqBody, _ := json.Marshal(FooRequest{Name: "Gin"})

	req, _ := http.NewRequest(http.MethodPost, "/foo", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp RespWithData[FooResponse]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, Gin", resp.Data.Message)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.NotEmpty(t, resp.Trace)
}

func TestWrapHandlerFiber(t *testing.T) {
	app := fiber.New()

	app.Post("/foo", WrapHandlerFiber(FooHandler))

	reqBody, _ := json.Marshal(FooRequest{Name: "Fiber"})

	req := httptest.NewRequest(http.MethodPost, "/foo", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	defer func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody RespWithData[FooResponse]
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Nil(t, err)
	assert.Equal(t, "Hello, Fiber", respBody.Data.Message)
	assert.Equal(t, http.StatusOK, respBody.Code)
	assert.NotEmpty(t, respBody.Trace)
}

func TestWrapHandler_Success(t *testing.T) {
	mock := &mockContext{
		inputJSON: TestReq{Name: "Wen"},
	}

	handlerFn := WrapHandler(func(_ Context, req TestReq) (TestResp, error) {
		assert.Equal(t, "Wen", req.Name)
		return TestResp{Greet: "Hello " + req.Name}, nil
	})

	err := handlerFn(mock)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, mock.outputCode)
	resp, ok := mock.outputValue.(RespWithData[TestResp])
	assert.True(t, ok)
	assert.Equal(t, "Hello Wen", resp.Data.Greet)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestWrapHandler_BindError(t *testing.T) {
	mock := &mockContext{
		errOnBind: errors.New("bad json"),
	}

	handlerFn := WrapHandler(func(_ Context, _ TestReq) (TestResp, error) {
		t.Fatal("should not be called")
		return TestResp{}, nil
	})

	err := handlerFn(mock)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusBadRequest, mock.outputCode)
	resp, ok := mock.outputValue.(BaseResp)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestWrapHandler_HandlerError(t *testing.T) {
	mock := &mockContext{
		inputJSON: TestReq{Name: "Wen"},
	}

	handlerFn := WrapHandler(func(_ Context, _ TestReq) (TestResp, error) {
		return TestResp{}, errors.New("internal error")
	})

	err := handlerFn(mock)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusInternalServerError, mock.outputCode)
	resp, ok := mock.outputValue.(BaseResp)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}
