// Package handler defines handler types.
package handler

// BaseResp is a base response.
type BaseResp struct {
	Code  int    `json:"code"`
	Trace string `json:"trace"`
}

// RespWithData is a response with data.
type RespWithData[T any] struct {
	BaseResp
	Data T `json:"data,omitempty"`
}
