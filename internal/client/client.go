package client

type Client struct {
	// JSON-RPC client
}

func NewClient() *Client {
	return &Client{}
}

type toolRequest struct {
	Tool string `json:"tool"`
}

func UseTool(tool string) error {
	return nil
}

func GetResource(path string) (string, error) {
	return "", nil
}
