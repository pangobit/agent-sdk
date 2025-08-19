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

type serverRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params"`
	ID      *json.RawMessage `json:"id"`
}

type Response struct {
	JSONRPC string           `json:"jsonrpc"`
	Result  any              `json:"result"`
	Error   *Error           `json:"error"`
	ID      *json.RawMessage `json:"id"`
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

	request serverRequest

	mutex   sync.Mutex
	seq     uint64
	pending map[uint64]*json.RawMessage
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	c.request.Method = ""
	c.request.Params = nil
	c.request.ID = nil

	if err := c.decoder.Decode(&c.request); err != nil {
		return ErrorParseError(err)
	}

	switch c.request.Method {
	case "":
		return ErrorMethodNotFound(nil)
	}

	if c.request.JSONRPC != jsonrpcVersion || c.request.JSONRPC == "" {
		return ErrorInvalidRequest(nil)
	}

	r.ServiceMethod = c.request.Method

	// JSON-RPC 2.0 allows the ID to be a "number", a string, or null.
	// Go's rpc package expects a uint64, so, like the old json1.0 rpc package,
	// we assign a uint64 and map it to the original for later.
	c.mutex.Lock()
	c.seq++
	c.pending[c.seq] = c.request.ID
	c.request.ID = nil
	r.Seq = c.seq
	c.mutex.Unlock()
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

var jsonNull = json.RawMessage([]byte("null"))

func (c *serverCodec) WriteResponse(r *rpc.Response, body any) error {
	c.mutex.Lock()

	b, ok := c.pending[r.Seq]
	if !ok {
		c.mutex.Unlock()
		return ErrorInvalidRequest(nil)
	}
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	if b == nil {
		b = &jsonNull
	}

	response := Response{
		JSONRPC: "2.0",
		ID:      b,
	}

	if r.Error == "" {
		response.Result = body
	} else {
		response.Result = r.Error
	}

	return c.encoder.Encode(response)
}

func (c *serverCodec) Close() error {
	return c.closer.Close()
}

type Connection interface {
	io.ReadWriteCloser
}
