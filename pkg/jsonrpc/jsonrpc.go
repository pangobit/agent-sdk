// Package jsonrpc implements JSON-RPC 2.0 specification.
// https://www.jsonrpc.org/specification
// Author: Pangobit, LLC
// License: MIT

package jsonrpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/rpc"
	"sync"
)

type Request struct {
	JSONRPC string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params"`
	ID      int              `json:"id"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	Result  any    `json:"result"`
	Error   *Error `json:"error"`
	ID      int    `json:"id"`
}

func validateResponse(response Response) error {
	if response.JSONRPC != "2.0" {
		return fmt.Errorf("invalid JSON-RPC version")
	}

	if response.Error != nil {
		return fmt.Errorf("error: %s", response.Error.Message)
	}
	return nil
}

// validateMethod validates that the method name follows JSON-RPC 2.0 specification.
// Method names that begin with "rpc." are reserved for rpc-internal methods and extensions.
func validateMethod(method string) error {
	if method == "" {
		return fmt.Errorf("method name cannot be empty")
	}

	// Check if method starts with "rpc." (case-sensitive)
	if len(method) >= 4 && method[:4] == "rpc." {
		return fmt.Errorf("method name '%s' is reserved for rpc-internal methods and extensions", method)
	}

	return nil
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorCode int

const (
	ErrorCodeParseError     ErrorCode = -32700
	ErrorCodeInvalidRequest ErrorCode = -32600
	ErrorCodeMethodNotFound ErrorCode = -32601
	ErrorCodeInvalidParams  ErrorCode = -32602
	ErrorCodeInternalError  ErrorCode = -32603
	ErrorCodeServerError    ErrorCode = -32000
)

type ErrorParseError error
type ErrorInvalidRequest error
type ErrorMethodNotFound error
type ErrorInvalidParams error
type ErrorInternalError error
type ErrorServerError error

type serverCodec struct {
	decoder *json.Decoder
	encoder *json.Encoder
	closer  io.Closer

	request Request

	lock sync.Mutex
	seq  uint64
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	if err := c.decoder.Decode(r); err != nil {
		return ErrorParseError(err)
	}

	switch c.request.Method {
	case "":
		return ErrorMethodNotFound(nil)
	}

	r.ServiceMethod = c.request.Method
	return nil
}

type nilParams struct{}

func (c *serverCodec) ReadRequestBody(body any) error {
	if body == nil {
		return ErrorInvalidRequest(nil)
	}

	var params [1]any
	if c.request.Params != nil {
		params[0] = body
	} else {
		params[0] = nilParams{}
	}

	return json.Unmarshal(*c.request.Params, &params)
}

func (c *serverCodec) WriteResponse(r *rpc.Response, body any) error {

	return nil
}

func (c *serverCodec) Close() error {
	return c.closer.Close()
}
