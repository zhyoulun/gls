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

func ReadUintBE(r io.Reader, n int) (uint32, error) {
	buf, err := ReadBytes(r, n)
	if err != nil {
		return 0, err
	}
	res := uint32(0)
	for i := 0; i < n; i++ {
		res = res<<8 + uint32(buf[i])
	}
	return res, nil
}

func ReadUintLE(r io.Reader, n int) (uint32, error) {
	buf, err := ReadBytes(r, n)
	if err != nil {
		return 0, err
	}
	res := uint32(0)
	for i := n - 1; i >= 0; i-- {
		res = res<<8 + uint32(buf[i])
	}
	return res, nil
}

func WriteUintBE(w io.Writer, v uint32, n int) error {
	for i := 0; i < n; i++ {
		b := byte(v>>uint32((n-i-1)<<3)) & 0xff
		if err := WriteByte(w, b); err != nil {
			return err
		}
	}
	return nil
}

func WriteUintLE(w io.Writer, v uint32, n int) error {
	for i := n - 1; i >= 0; i-- {
		b := byte(v>>uint32((n-i-1)<<3)) & 0xff
		if err := WriteByte(w, b); err != nil {
			return err
		}
	}
	return nil
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
type WriteBufferWithMaxCapacity struct {
	data     []byte
	capacity int
}

func NewWriteBufferWithMaxCapacity(capacity int) (*WriteBufferWithMaxCapacity, error) {
	return &WriteBufferWithMaxCapacity{
		data:     make([]byte, 0, capacity),
		capacity: capacity,
	}, nil
}

func (b *WriteBufferWithMaxCapacity) Write(buf []byte) (int, error) {
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

func (b *WriteBufferWithMaxCapacity) Bytes() []byte {
	return b.data
}

//used for unit test
type ReadBuffer struct {
	data        []byte
	toReadIndex int
}

func NewReadBuffer(initData []byte) (*ReadBuffer, error) {
	data := make([]byte, 0)
	data = append(data, initData...)
	return &ReadBuffer{
		data:        data,
		toReadIndex: 0,
	}, nil
}

func (b *ReadBuffer) Write(p []byte) {
	b.data = append(b.data, p...)
}

func (b *ReadBuffer) Read(p []byte) (int, error) {
	if b.toReadIndex == len(b.data) {
		return 0, fmt.Errorf("eof")
	}
	if len(b.data)-b.toReadIndex <= len(p) {
		copy(p[:len(b.data)-b.toReadIndex], b.data[b.toReadIndex:])
		l := len(b.data) - b.toReadIndex
		b.toReadIndex = len(b.data)
		return l, nil
	} else {
		copy(p, b.data[b.toReadIndex:b.toReadIndex+len(p)])
		l := len(p)
		b.toReadIndex += len(p)
		return l, nil
	}
}

func (b *ReadBuffer) Peek(n int) ([]byte, error) {
	if len(b.data)-b.toReadIndex >= n {
		return b.data[b.toReadIndex : b.toReadIndex+n], nil
	} else {
		return nil, fmt.Errorf("not enough data in buffer")
	}
}
