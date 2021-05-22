package rtmp

import (
	"github.com/stretchr/testify/assert"
	"github.com/zhyoulun/gls/src/utils"
	"github.com/zhyoulun/gls/src/utils/debug"
	"testing"
)

func Test_chunkStream_readChunkFmt0(t *testing.T) {
	_ = debug.Init()
	header := &chunkBasicHeader{
		chunkStreamID: 1,
		fmt:           fmt0,
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00})                               //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08})                               //message length
		buf.Write([]byte{0x00})                                           //message type id
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00})                         //message stream id
		buf.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) //data

		err := cs.readChunk(buf, 4)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), cs.clock)
		assert.Equal(t, uint32(8), cs.messageLength)
		assert.Equal(t, uint8(0), cs.messageTypeID)
		assert.Equal(t, uint32(0), cs.messageStreamID)
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}, cs.data)
		assert.Equal(t, uint32(4), cs.dataIndex)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0xff, 0xff, 0xff})                               //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08})                               //message length
		buf.Write([]byte{0x00})                                           //message type id
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00})                         //message stream id
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00})                         //timestamp
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //data

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), cs.clock)
		assert.Equal(t, uint32(8), cs.messageLength)
		assert.Equal(t, uint8(0), cs.messageTypeID)
		assert.Equal(t, uint32(0), cs.messageStreamID)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00}) //timestamp

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00}) //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08}) //message length

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00}) //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08}) //message length
		buf.Write([]byte{0x00})             //message type id

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00})       //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08})       //message length
		buf.Write([]byte{0x00})                   //message type id
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) //message stream id

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00})                         //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08})                         //message length
		buf.Write([]byte{0x00})                                     //message type id
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00})                   //message stream id
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //data

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0xff, 0xff, 0xff})       //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08})       //message length
		buf.Write([]byte{0x00})                   //message type id
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) //message stream id

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
}

func Test_chunkStream_readChunkFmt1(t *testing.T) {
	_ = debug.Init()
	header := &chunkBasicHeader{
		chunkStreamID: 1,
		fmt:           fmt1,
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00})                               //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08})                               //message length
		buf.Write([]byte{0x00})                                           //message type id
		buf.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) //data

		err := cs.readChunk(buf, 4)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), cs.clock)
		assert.Equal(t, uint32(8), cs.messageLength)
		assert.Equal(t, uint8(0), cs.messageTypeID)
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}, cs.data)
		assert.Equal(t, uint32(4), cs.dataIndex)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0xff, 0xff, 0xff})                               //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08})                               //message length
		buf.Write([]byte{0x00})                                           //message type id
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00})                         //timestamp
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //data

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), cs.clock)
		assert.Equal(t, uint32(8), cs.messageLength)
		assert.Equal(t, uint8(0), cs.messageTypeID)
		assert.Equal(t, uint32(0), cs.messageStreamID)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00}) //timestamp

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00}) //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08}) //message length

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00}) //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08}) //message length
		buf.Write([]byte{0x00})             //message type id

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00})                         //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08})                         //message length
		buf.Write([]byte{0x00})                                     //message type id
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //data

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0xff, 0xff, 0xff}) //timestamp
		buf.Write([]byte{0x00, 0x00, 0x08}) //message length
		buf.Write([]byte{0x00})             //message type id

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
}

func Test_chunkStream_readChunkFmt2(t *testing.T) {
	_ = debug.Init()
	header := &chunkBasicHeader{
		chunkStreamID: 1,
		fmt:           fmt2,
	}
	{
		cs, _ := newChunkStreamForRead(header)
		cs.messageLength = 8
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00})                               //timestamp
		buf.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) //data

		err := cs.readChunk(buf, 4)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), cs.clock)
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}, cs.data)
		assert.Equal(t, uint32(4), cs.dataIndex)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		cs.messageLength = 8
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0xff, 0xff, 0xff})                               //timestamp
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00})                         //timestamp
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //data

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), cs.clock)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		cs.messageLength = 8
		buf, _ := utils.NewReadBuffer([]byte{})

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		cs.messageLength = 8
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00}) //timestamp

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		cs.messageLength = 8
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00})                         //timestamp
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //data

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		cs.messageLength = 8
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0xff, 0xff, 0xff}) //timestamp

		err := cs.readChunk(buf, defaultMaximumChunkSize)
		assert.Error(t, err)
	}
}

func Test_chunkStream_readChunkFmt3(t *testing.T) {
	_ = debug.Init()
	header := &chunkBasicHeader{
		chunkStreamID: 1,
		fmt:           fmt3,
	}
	{
		cs, _ := newChunkStreamForRead(header)
		cs.messageLength = 8
		cs.tmp.extended = true
		cs.dataIndex = 8
		//cs.initData(cs.messageLength)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x0a, 0x00, 0x0b, 0x00})                         //extended timestamp
		buf.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) //data

		err := cs.readChunk(buf, 4)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x0a000b00), cs.clock)
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}, cs.data)
		assert.Equal(t, uint32(4), cs.dataIndex)
		assert.Equal(t, uint32(0x0a000b00), cs.tmp.extendedTimestamp)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		cs.messageLength = 8
		cs.tmp.extended = false
		cs.dataIndex = 8
		//cs.initData(cs.messageLength)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) //data

		err := cs.readChunk(buf, 4)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), cs.clock)
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}, cs.data)
		assert.Equal(t, uint32(4), cs.dataIndex)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		cs.messageLength = 8
		cs.tmp.extended = false
		cs.dataIndex = 8
		cs.fmt = fmt1
		//cs.initData(cs.messageLength)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) //data

		err := cs.readChunk(buf, 4)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), cs.clock)
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}, cs.data)
		assert.Equal(t, uint32(4), cs.dataIndex)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		cs.messageLength = 8
		cs.tmp.extended = true
		cs.dataIndex = 8
		cs.fmt = fmt1
		//cs.initData(cs.messageLength)
		buf, _ := utils.NewReadBuffer([]byte{})
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00})                         //extended timestamp
		buf.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) //data

		err := cs.readChunk(buf, 4)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), cs.clock)
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}, cs.data)
		assert.Equal(t, uint32(4), cs.dataIndex)
	}
	{
		cs, _ := newChunkStreamForRead(header)
		cs.messageLength = 8
		cs.tmp.extended = true
		cs.dataIndex = 8
		cs.fmt = fmt1
		//cs.initData(cs.messageLength)
		buf, _ := utils.NewReadBuffer([]byte{})
		//buf.Write([]byte{0x00, 0x00, 0x00, 0x00})                         //extended timestamp
		//buf.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) //data

		err := cs.readChunk(buf, 4)
		assert.Error(t, err)
	}
}
