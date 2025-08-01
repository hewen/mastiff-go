// Package handler defines handler types.
package handler

// BaseResp is a base response.
type BaseResp struct {
	Trace string `json:"trace"`
	Code  int    `json:"code"`
}

// RespWithData is a response with data.
type RespWithData[T any] struct {
	Data T `json:"data,omitempty"`
	BaseResp
}
