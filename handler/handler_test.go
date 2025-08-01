package handler

import (
	"context"
	"errors"
)

type TestReq struct {
	Name string `json:"name"`
}

type TestResp struct {
	Greet string `json:"greet"`
}

type mockContext struct {
	errOnBind   error
	inputJSON   any
	outputValue any
	setValues   map[string]any
	outputCode  int
}

func (m *mockContext) BindJSON(obj any) error {
	if m.errOnBind != nil {
		return m.errOnBind
	}
	ptr, ok := obj.(*TestReq)
	if !ok {
		return errors.New("invalid type")
	}
	*ptr = m.inputJSON.(TestReq)
	return nil
}

func (m *mockContext) RequestContext() context.Context {
	return context.Background()
}

func (m *mockContext) JSON(code int, value any) error {
	m.outputCode = code
	m.outputValue = value
	return nil
}

func (m *mockContext) Set(key string, val any) {
	if m.setValues == nil {
		m.setValues = make(map[string]any)
	}
	m.setValues[key] = val
}
