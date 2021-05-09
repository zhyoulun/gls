package utils

import (
	"bufio"
	"io"
	"net"
	"time"
)

type BufferedConn struct {
	conn net.Conn
	rw   *bufio.ReadWriter
}

func NewBufferedConn(conn net.Conn, bufferSize int) *BufferedConn {
	return &BufferedConn{
		conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReaderSize(conn, bufferSize), bufio.NewWriterSize(conn, bufferSize)),
	}
}

func (bc *BufferedConn) Read(b []byte) (n int, err error) {
	return io.ReadAtLeast(bc.rw, b, len(b))
}

func (bc *BufferedConn) Write(b []byte) (n int, err error) {
	return bc.rw.Write(b)
}

func (bc *BufferedConn) Close() error {
	return bc.conn.Close()
}

func (bc *BufferedConn) LocalAddr() net.Addr {
	return bc.conn.LocalAddr()
}

func (bc *BufferedConn) RemoteAddr() net.Addr {
	return bc.conn.RemoteAddr()
}

func (bc *BufferedConn) SetDeadline(t time.Time) error {
	return bc.conn.SetDeadline(t)
}

func (bc *BufferedConn) SetReadDeadline(t time.Time) error {
	return bc.conn.SetReadDeadline(t)
}

func (bc *BufferedConn) SetWriteDeadline(t time.Time) error {
	return bc.conn.SetWriteDeadline(t)
}

func (bc *BufferedConn) Flush() error {
	if bc.rw.Writer.Buffered() == 0 {
		return nil
	}
	return bc.rw.Flush()
}
