package tlc

import (
	"bufio"
	"encoding/gob"
	"io"
	"net"
	"net/rpc"
)

// This codec uses net/rpc's ordinary gob header/body format. Its only extra
// responsibility is tracking accepted requests until their response is flushed;
// ServeConn does not expose that point to a process owner performing shutdown.
type distributedServerCodec struct {
	server *DistributedRPCServer
	conn   net.Conn
	decode *gob.Decoder
	encode *gob.Encoder
	buffer *bufio.Writer
}

func newDistributedServerCodec(server *DistributedRPCServer, conn net.Conn) *distributedServerCodec {
	buffer := bufio.NewWriter(conn)
	return &distributedServerCodec{server: server, conn: conn, decode: gob.NewDecoder(conn), encode: gob.NewEncoder(buffer), buffer: buffer}
}

func (c *distributedServerCodec) ReadRequestHeader(request *rpc.Request) error {
	if err := c.decode.Decode(request); err != nil {
		return err
	}
	c.server.mu.Lock()
	defer c.server.mu.Unlock()
	if c.server.closed {
		return io.EOF
	}
	// Close marks closed under this same lock before Wait, so Add can never
	// race the transition from zero pending replies to shutdown completion.
	c.server.replies.Add(1)
	return nil
}

func (c *distributedServerCodec) ReadRequestBody(body any) error {
	return c.decode.Decode(body)
}

func (c *distributedServerCodec) WriteResponse(response *rpc.Response, body any) error {
	defer c.server.replies.Done()
	if err := c.encode.Encode(response); err != nil {
		_ = c.Close()
		return err
	}
	if err := c.encode.Encode(body); err != nil {
		_ = c.Close()
		return err
	}
	return c.buffer.Flush()
}

func (c *distributedServerCodec) Close() error { return c.conn.Close() }

var _ rpc.ServerCodec = (*distributedServerCodec)(nil)
