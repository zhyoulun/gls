package rtmp

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/zhyoulun/gls/src/utils"
	"os"
	"testing"
)

func Test_chunkBasicHeader_Read(t *testing.T) {
	{
		header, _ := newChunkBasicHeaderForRead()
		buf := bytes.NewBuffer([]byte{0x00, 0x00})
		err := header.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, fmt0, header.fmt)
		assert.Equal(t, uint32(64), header.chunkStreamID)
		fmt.Fprintf(os.Stdout, "%s", header)
	}
	{
		header, _ := newChunkBasicHeaderForRead()
		buf := bytes.NewBuffer([]byte{0x01, 0x00, 0x00})
		err := header.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, fmt0, header.fmt)
		assert.Equal(t, uint32(64), header.chunkStreamID)
	}
	{
		header, _ := newChunkBasicHeaderForRead()
		buf := bytes.NewBuffer([]byte{0x02})
		err := header.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, fmt0, header.fmt)
		assert.Equal(t, uint32(2), header.chunkStreamID)
	}
	{
		header, _ := newChunkBasicHeaderForRead()
		buf := bytes.NewBuffer([]byte{0x03})
		err := header.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, fmt0, header.fmt)
		assert.Equal(t, uint32(3), header.chunkStreamID)
	}
	{
		header, _ := newChunkBasicHeaderForRead()
		buf := bytes.NewBuffer([]byte{})
		err := header.Read(buf)
		assert.Error(t, err)
	}
	{
		header, _ := newChunkBasicHeaderForRead()
		buf := bytes.NewBuffer([]byte{0x00})
		err := header.Read(buf)
		assert.Error(t, err)
	}
	{
		header, _ := newChunkBasicHeaderForRead()
		buf := bytes.NewBuffer([]byte{0x01})
		err := header.Read(buf)
		assert.Error(t, err)
	}
}

func Test_chunkBasicHeader_Write(t *testing.T) {
	{
		header, _ := newChunkBasicHeaderForWrite(0, fmt0)
		buf := &bytes.Buffer{}
		err := header.Write(buf)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x00}, buf.Bytes())
	}
	{
		header, _ := newChunkBasicHeaderForWrite(64, fmt0)
		buf := &bytes.Buffer{}
		err := header.Write(buf)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x00, 0x00}, buf.Bytes())
	}
	{
		header, _ := newChunkBasicHeaderForWrite(0x01f4+64, fmt0)
		buf := &bytes.Buffer{}
		err := header.Write(buf)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x01, 0xf4, 0x01}, buf.Bytes())
	}
	{
		header, _ := newChunkBasicHeaderForWrite(0, fmt0)
		buf, _ := utils.NewWriteBufferWithMaxCapacity(0)
		err := header.Write(buf)
		assert.Error(t, err)
	}
	{
		header, _ := newChunkBasicHeaderForWrite(64, fmt0)
		buf, _ := utils.NewWriteBufferWithMaxCapacity(0)
		err := header.Write(buf)
		assert.Error(t, err)
	}
	{
		header, _ := newChunkBasicHeaderForWrite(64, fmt0)
		buf, _ := utils.NewWriteBufferWithMaxCapacity(1)
		err := header.Write(buf)
		assert.Error(t, err)
	}
	{
		header, _ := newChunkBasicHeaderForWrite(0x01f4+64, fmt0)
		buf, _ := utils.NewWriteBufferWithMaxCapacity(0)
		err := header.Write(buf)
		assert.Error(t, err)
	}
	{
		header, _ := newChunkBasicHeaderForWrite(0x01f4+64, fmt0)
		buf, _ := utils.NewWriteBufferWithMaxCapacity(1)
		err := header.Write(buf)
		assert.Error(t, err)
	}
}
