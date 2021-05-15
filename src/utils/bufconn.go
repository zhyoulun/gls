package utils

import (
	"bufio"
	"io"
	"net"
)

type ReaderPeeker interface {
	io.Reader
	Peek(n int) ([]byte, error)
}

type PeekerConn interface {
	net.Conn
	Peek(n int) ([]byte, error)
}

type BufferedConn struct {
	net.Conn // So that most methods are embedded
	r        *bufio.Reader
}

func NewBufferedConn(conn net.Conn, bufferSize int) *BufferedConn {
	return &BufferedConn{conn, bufio.NewReaderSize(conn, bufferSize)}
}

func (bc *BufferedConn) Peek(n int) ([]byte, error) {
	return bc.r.Peek(n)
}
