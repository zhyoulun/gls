package amf

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_readBytes(t *testing.T) {
	{
		r := strings.NewReader("abc")
		buf, err := readBytes(r, 3)
		assert.Equal(t, buf, []byte(`abc`))
		assert.NoError(t, err)
	}
	{
		r := strings.NewReader("abc")
		buf, err := readBytes(r, 4)
		assert.Nil(t, buf)
		assert.Error(t, err)
	}
}

func Test_readByte(t *testing.T) {
	{
		r := strings.NewReader("abc")
		b, err := readByte(r)
		assert.Equal(t, b, byte('a'))
		assert.NoError(t, err)
	}
	{
		r := strings.NewReader("")
		b, err := readByte(r)
		assert.Equal(t, b, byte(0))
		assert.Error(t, err)
	}
}

func Test_writeBytes(t *testing.T) {
	{
		buf := &bytes.Buffer{}
		err := writeBytes(buf, []byte(`abc`))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`abc`), buf.Bytes())
	}
	{
		buf, _ := newBufferWithMaxCapacity(3)
		err := writeBytes(buf, []byte(`abcd`))
		assert.Error(t, err)
		assert.Equal(t, []byte(``), buf.data)
	}
	{
		buf, _ := newBufferWithMaxCapacity(12)
		err := writeBytes(buf, []byte(`abcdef`))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`abcdef`), buf.data)
	}
}

func Test_writeByte(t *testing.T) {
	{
		{
			buf := &bytes.Buffer{}
			err := writeByte(buf, 'a')
			assert.NoError(t, err)
			assert.Equal(t, []byte(`a`), buf.Bytes())
		}
	}
}
