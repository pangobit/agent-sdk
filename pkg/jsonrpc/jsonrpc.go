package jsonrpc

import "fmt"

type Request struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      int    `json:"id"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	Result  any    `json:"result"`
	Error   *Error `json:"error"`
	ID      int    `json:"id"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
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
