package utils

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_readBytes(t *testing.T) {
	{
		r := strings.NewReader("abc")
		buf, err := ReadBytes(r, 3)
		assert.Equal(t, buf, []byte(`abc`))
		assert.NoError(t, err)
	}
	{
		r := strings.NewReader("abc")
		buf, err := ReadBytes(r, 4)
		assert.Nil(t, buf)
		assert.Error(t, err)
	}
}

func Test_readByte(t *testing.T) {
	{
		r := strings.NewReader("abc")
		b, err := ReadByte(r)
		assert.Equal(t, b, byte('a'))
		assert.NoError(t, err)
	}
	{
		r := strings.NewReader("")
		b, err := ReadByte(r)
		assert.Equal(t, b, byte(0))
		assert.Error(t, err)
	}
}

func Test_writeBytes(t *testing.T) {
	{
		buf := &bytes.Buffer{}
		err := WriteBytes(buf, []byte(`abc`))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`abc`), buf.Bytes())
	}
	{
		buf, _ := NewBufferWithMaxCapacity(3)
		err := WriteBytes(buf, []byte(`abcd`))
		assert.Error(t, err)
		assert.Equal(t, []byte(``), buf.data)
	}
	{
		buf, _ := NewBufferWithMaxCapacity(12)
		err := WriteBytes(buf, []byte(`abcdef`))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`abcdef`), buf.data)
	}
}

func Test_writeByte(t *testing.T) {
	{
		{
			buf := &bytes.Buffer{}
			err := WriteByte(buf, 'a')
			assert.NoError(t, err)
			assert.Equal(t, []byte(`a`), buf.Bytes())
		}
	}
}

func TestReadUintBE(t *testing.T) {
	{
		buf := bytes.NewBuffer([]byte{0x01})
		n, err := ReadUintBE(buf, 1)
		assert.NoError(t, err)
		assert.Equal(t, uint32(1), n)
	}
	{
		buf := bytes.NewBuffer([]byte{0x01, 0x02})
		n, err := ReadUintBE(buf, 2)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x0102), n)
	}
	{
		buf := &bytes.Buffer{}
		_, err := ReadUintBE(buf, 2)
		assert.Error(t, err)
	}
}

func TestWriteUintBE(t *testing.T) {
	{
		buf := &bytes.Buffer{}
		err := WriteUintBE(buf, 0x01, 1)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x01}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		err := WriteUintBE(buf, 0x0102, 2)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x01, 0x02}, buf.Bytes())
	}
	{
		buf, _ := NewBufferWithMaxCapacity(1)
		err := WriteUintBE(buf, 0x0102, 2)
		assert.Error(t, err)
	}
}

func TestReadUintLE(t *testing.T) {
	{
		buf := bytes.NewBuffer([]byte{0x01})
		n, err := ReadUintLE(buf, 1)
		assert.NoError(t, err)
		assert.Equal(t, uint32(1), n)
	}
	{
		buf := bytes.NewBuffer([]byte{0x01, 0x02})
		n, err := ReadUintLE(buf, 2)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x0201), n)
	}
	{
		buf := &bytes.Buffer{}
		_, err := ReadUintLE(buf, 2)
		assert.Error(t, err)
	}
}

func TestWriteUintLE(t *testing.T) {
	{
		buf := &bytes.Buffer{}
		err := WriteUintLE(buf, 0x01, 1)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x01}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		err := WriteUintLE(buf, 0x0102, 2)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x02, 0x01}, buf.Bytes())
	}
	{
		buf, _ := NewBufferWithMaxCapacity(1)
		err := WriteUintLE(buf, 0x0102, 2)
		assert.Error(t, err)
	}
}
