// Adapted from the stdlib json-rpc 1.0 library.

package jsonrpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"sync"
)

const jsonrpcVersion = "2.0"

type clientCodec struct {
	decoder *json.Decoder
	encoder *json.Encoder
	closer  io.Closer

	request  clientRequest
	response clientResponse

	mutex   sync.Mutex
	pending map[uint64]string
}

type clientRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  [1]any `json:"params"`
	ID      uint64 `json:"id"`
}

func (c *clientCodec) WriteRequest(r *rpc.Request, param any) error {
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()

	c.request.JSONRPC = jsonrpcVersion
	c.request.Method = r.ServiceMethod
	c.request.Params[0] = param
	c.request.ID = r.Seq

	return c.encoder.Encode(&c.request)
}

type clientResponse struct {
	Result *json.RawMessage `json:"result"`
	Error  any              `json:"error"`
	ID     uint64           `json:"id"`
}

func reset(r clientResponse) clientResponse {
	req := r
	r.ID = 0
	r.Result = nil
	r.Error = nil
	return req
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	c.response = reset(c.response)
	if err := c.decoder.Decode(&c.response); err != nil {
		return err
	}

	c.mutex.Lock()
	r.ServiceMethod = c.pending[c.response.ID]
	delete(c.pending, c.response.ID)
	c.mutex.Unlock()

	r.Error = ""
	r.Seq = c.response.ID
	if c.response.Error != nil || c.response.Result == nil {
		x, ok := c.response.Error.(string)
		if !ok {
			return fmt.Errorf("invalid error: %v", c.response.Error)
		}
		if x != "" {
			r.Error = "unspecified error"
		}
		r.Error = x
	}

	return nil
}

func (c *clientCodec) ReadResponseBody(body any) error {
	if body == nil {
		return nil
	}

	return json.Unmarshal(*c.response.Result, body)
}

func (c *clientCodec) Close() error {
	return c.closer.Close()
}

func NewClient(conn Connection) *rpc.Client {
	return rpc.NewClientWithCodec(&clientCodec{
		decoder: json.NewDecoder(conn),
		encoder: json.NewEncoder(conn),
		closer:  conn,
		pending: make(map[uint64]string),
	})
}

func Dial(network, address string) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}
