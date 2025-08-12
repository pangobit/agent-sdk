package stdio

import (
	"io"
	"os"
)

type Connection struct {
	reader io.Reader
	writer io.Writer
}

func Dial() *Connection {
	return &Connection{
		reader: os.Stdin,
		writer: os.Stdout,
	}
}

func (c *Connection) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}

func (c *Connection) Write(p []byte) (n int, err error) {
	return c.writer.Write(p)
}

func (c *Connection) Close() error {
	return nil
}
