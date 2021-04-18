package utils

import (
	"fmt"
	"io"
)

func ReadByte(r io.Reader) (byte, error) {
	buf, err := ReadBytes(r, 1)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func ReadBytes(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n) //todo 优化
	p := buf
	for {
		m, err := r.Read(p)
		if err != nil {
			return nil, err
		}
		if m == n {
			break
		}
		n -= m
		p = p[m:]
	}
	return buf, nil
}

func WriteByte(w io.Writer, b byte) error {
	return WriteBytes(w, []byte{b})
}

func WriteBytes(w io.Writer, buf []byte) error {
	for {
		n, err := w.Write(buf)
		if err != nil {
			return err
		}
		if n == len(buf) {
			return nil
		}
		buf = buf[n:]
	}
}

//used for unit test
type BufferWithMaxCapacity struct {
	data     []byte
	capacity int
}

func NewBufferWithMaxCapacity(capacity int) (*BufferWithMaxCapacity, error) {
	return &BufferWithMaxCapacity{
		data:     make([]byte, 0, capacity),
		capacity: capacity,
	}, nil
}

func (b *BufferWithMaxCapacity) Write(buf []byte) (int, error) {
	if len(b.data)+len(buf) > b.capacity {
		return 0, fmt.Errorf("overflow")
	}
	if len(buf) > 5 {
		b.data = append(b.data, buf[0:5]...)
		return 5, nil
	} else {
		b.data = append(b.data, buf...)
		return len(buf), nil
	}
}

func (b *BufferWithMaxCapacity) Bytes() []byte {
	return b.data
}
