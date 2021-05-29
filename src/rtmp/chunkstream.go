package rtmp

import (
	"fmt"
	"github.com/zhyoulun/gls/src/core"
	"github.com/zhyoulun/gls/src/utils"
	"github.com/zhyoulun/gls/src/utils/debug"
	"io"
)

type chunkStream struct {
	fmt             Fmt
	chunkStreamID   uint32
	clock           uint32
	messageLength   uint32
	messageTypeID   uint8
	messageStreamID uint32
	tmp             chunkStreamTmp
	data            []byte //chunk data
	dataIndex       uint32 //待读取的位置
}

func (cs *chunkStream) toChunkCsvHeader() string {
	return fmt.Sprintf("fmt,currentFmt,chunkStreamID,timestamp,messageLength,messageTypeID,messageStreamID,extended,timestampDelta,readLength\n")
}

func (cs *chunkStream) toChunkCsvLine() string {
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%t,%d,%d\n", cs.fmt, cs.tmp.currentFmt,
		cs.chunkStreamID, cs.clock, cs.messageLength, cs.messageTypeID, cs.messageStreamID,
		cs.tmp.extended, cs.tmp.timestampDelta, cs.dataIndex)
}

func newChunkStreamForRead(basicHeader *chunkBasicHeader) (*chunkStream, error) {
	return &chunkStream{
		chunkStreamID: basicHeader.chunkStreamID,
		tmp: chunkStreamTmp{
			currentFmt: basicHeader.fmt,
		},
	}, nil
}

//chunk stream中的临时变量，对chunk stream本身没作用，主要是用来基于若干个chunk组合出一个完整的message based on chunk stream
type chunkStreamTmp struct {
	currentFmt         Fmt    //chunk type
	extended           bool   //是否需要读取extend timestamp，用于fmt3 chunk的读取场景
	firstChunkReadDone bool   //chunkStream中，第一个chunk读取完成
	timestamp          uint32 //used for read storage
	timestampDelta     uint32 //used for read storage
	extendedTimestamp  uint32 //used for read storage
	//extended          bool
	//extendedTimestamp uint32
	//readDone   bool
	//leftToRead uint32
}

func (cs *chunkStream) setBasicHeader(header *chunkBasicHeader) {
	cs.tmp.currentFmt = header.fmt
	cs.chunkStreamID = header.chunkStreamID
}

func (cs *chunkStream) String() string {
	return fmt.Sprintf("timestamp: %d, messageLength: %d, messageTypeID: %d, messageStreamID: %d, dataIndex: %d",
		cs.clock, cs.messageLength, cs.messageTypeID, cs.messageStreamID, cs.dataIndex)
}

func (cs *chunkStream) gotOneMessage() bool {
	return cs.messageLength == cs.dataIndex
}

func (cs *chunkStream) readChunk(r utils.ReadPeeker, chunkSize uint32) error {
	switch cs.tmp.currentFmt {
	case fmt0: //11B, this type must be used at the start of a chunk stream
		cs.fmt = fmt0
		var err error
		if cs.tmp.timestamp, err = utils.ReadUintBE(r, 3); err != nil {
			return err
		}
		//cs.messageLength
		if cs.messageLength, err = utils.ReadUintBE(r, 3); err != nil {
			return err
		}
		//cs.messageTypeID
		if messageTypeID, err := utils.ReadUintBE(r, 1); err != nil {
			return err
		} else {
			cs.messageTypeID = uint8(messageTypeID)
		}
		//cs.messageStreamID
		//typically, all messages in the same chunk stream will come from the same message stream.
		//while it is possible to multiplex separate message streams into the same chunk stream,
		//this defeats the benefits of the header compression.
		//However, if one message stream is closed and another one subsequently opened,
		//there is no reason an existing chunk stream cannot be reused by sending a new type-0 chunk.
		if cs.messageStreamID, err = utils.ReadUintLE(r, 4); err != nil {
			return err
		}
		//cs.timestamp
		if cs.tmp.timestamp == max3BTimestamp {
			var extendedTimestamp uint32
			if extendedTimestamp, err = utils.ReadUintBE(r, 4); err != nil {
				return err
			}
			cs.clock = extendedTimestamp
			cs.tmp.extended = true
			cs.tmp.timestampDelta = 0
		} else {
			cs.clock = cs.tmp.timestamp
			cs.tmp.extended = false
			cs.tmp.timestampDelta = 0
		}
		//cs.data init
		cs.initData(cs.messageLength)
	case fmt1: //7B, stream with variable-sized message(for example: many video formats) should use this format for the first chunk of each new message after the first(//todo first what??)
		cs.fmt = fmt1
		var err error
		//for a type-1 or type-2 chunk, the difference between the previous chunk's timestamp and the current chunk's timestamp is sent here.
		if cs.tmp.timestampDelta, err = utils.ReadUintBE(r, 3); err != nil {
			return err
		}
		//cs.messageLength
		if cs.messageLength, err = utils.ReadUintBE(r, 3); err != nil {
			return err
		}
		//cs.messageTypeID
		if messageTypeID, err := utils.ReadUintBE(r, 1); err != nil {
			return err
		} else {
			cs.messageTypeID = uint8(messageTypeID)
		}
		//cs.timestamp
		if cs.tmp.timestampDelta == max3BTimestamp {
			var extendedTimestamp uint32
			if extendedTimestamp, err = utils.ReadUintBE(r, 4); err != nil {
				return err
			}
			cs.clock += extendedTimestamp
			cs.tmp.extended = true
			cs.tmp.timestampDelta = extendedTimestamp
		} else {
			cs.clock += cs.tmp.timestampDelta
			cs.tmp.extended = false
			//cs.tmp.timestampDelta = cs.tmp.timestampDelta
		}
		//cs.data init
		cs.initData(cs.messageLength)
	case fmt2: //3B, stream with constant-sized messages(for example: some audio and data formats) should use this format for the first chunk of each message after the first
		cs.fmt = fmt2
		var err error
		if cs.tmp.timestampDelta, err = utils.ReadUintBE(r, 3); err != nil {
			return err
		}
		//cs.timestamp
		if cs.tmp.timestampDelta == max3BTimestamp {
			var extendedTimestamp uint32
			if extendedTimestamp, err = utils.ReadUintBE(r, 4); err != nil {
				return err
			}
			cs.clock += extendedTimestamp
			cs.tmp.extended = true
			cs.tmp.timestampDelta = extendedTimestamp
		} else {
			cs.clock += cs.tmp.timestampDelta
			cs.tmp.extended = false
			//cs.tmp.timestampDelta = cs.tmp.timestampDelta
		}
		//cs.data init
		cs.initData(cs.messageLength)
	case fmt3: //0B; fmt3不可能是ChunkStream的第一个chunk
		//chunks of this type take values from the preceding chunk for the same chunk stream id.
		//when a single message is split into chunks, all chunks of a message except the first one SHOULD use this type. Refer to Example 2
		//a stream consisting of messages of exactly the same size, stream id and spacing in time SHOULD use the this type for all chunks after a chunk of type 2. Refer to Example 1
		//if the delta between the first message and the second message is same as the timestamp of the first message, then a chunk of type 3 could immediately follow the chunk of type 0
		//		as there is no need for a chunk of type 2 to register the delta
		//if a type 3 chunk follows a type 0 chunk, then the timestamp delta for this type 3 chunk is the same as the timestamp of the type 0 chunk

		//read extended timestamp
		if cs.tmp.extended {
			var err error
			if cs.tmp.extendedTimestamp, err = utils.ReadUintBE(r, 4); err != nil {
				return err
			}
		} else {
			//nothing changed
		}
		//chunk type 3 can be used in two different ways. the first is to specify the continuation of a message.
		//  the second is to specify the beginning of a new message whose header can be derived from the existing state data.
		if cs.dataIndex == cs.messageLength { //example 1
			//set cs.clock
			if cs.fmt == fmt0 {
				if cs.tmp.extended {
					cs.clock = cs.tmp.extendedTimestamp
				} else {
					//nothing changed
				}
			} else {
				if cs.tmp.extended {
					cs.clock += cs.tmp.extendedTimestamp
				} else {
					cs.clock += cs.tmp.timestampDelta
				}
			}
			cs.initData(cs.messageLength)
		} else {
			//nothing changed
		}

	default:
		return core.ErrorImpossible
	}

	//cs.data read
	//一般读chunkSize个字节数，最后一次可能比chunkSize小
	//message length is for a type-0 or type-1 chunk.
	//Note this is generally not the same as the length of the chunk payload.
	//the chunk payload length is the maximum chunk size for all but the last chunk,
	//		and the remainder(which may be the entire length, for small messages) for the last chunk//todo 什么意思??
	readLength := cs.messageLength - cs.dataIndex //example 2
	if readLength > chunkSize {
		readLength = chunkSize
	}
	if buf, err := utils.ReadBytes(r, int(readLength)); err != nil { //todo 这里的强制类型转换需要确认下是否有风险
		return err
	} else {
		copy(cs.data[cs.dataIndex:cs.dataIndex+readLength], buf)
		cs.dataIndex += readLength
	}

	debug.Csv.Write(&debug.Message{
		FileName:   "chunk.csv",
		HeaderLine: cs.toChunkCsvHeader(),
		BodyLine:   cs.toChunkCsvLine(),
	})

	return nil
}

func (cs *chunkStream) initData(messageLength uint32) {
	cs.data = make([]byte, messageLength)
	cs.dataIndex = 0
	cs.tmp.firstChunkReadDone = true
}

func (cs *chunkStream) writeChunk(w io.Writer, chunkSize uint32) error {
	numChunks := cs.messageLength / chunkSize
	for i := uint32(0); i <= numChunks; i++ {
		//debug
		if numChunks > 0 {
			_ = chunkSize
		}
		var f Fmt
		if i == 0 {
			f = fmt0
		} else {
			f = fmt3
		}
		if basicHeader, err := newChunkBasicHeaderForWrite(cs.chunkStreamID, f); err != nil {
			return err
		} else {
			if err := basicHeader.Write(w); err != nil {
				return err
			}
		}
		//chunk message header
		if f == fmt3 {
			if cs.clock > 0xffffff {
				//todo ??为什么要删掉
				//if err := utils.WriteUintBE(w, 0xffffff, 3); err != nil {
				//	return err
				//}
				if err := utils.WriteUintBE(w, cs.clock, 4); err != nil {
					return err
				}
			} else {
				//todo ??为什么要删掉
				//if err := utils.WriteUintBE(w, cs.timestamp, 3); err != nil {
				//	return err
				//}
			}
		} else if f == fmt0 {
			if cs.clock > 0xffffff {
				if err := utils.WriteUintBE(w, 0xffffff, 3); err != nil {
					return err
				}
				if err := utils.WriteUintBE(w, cs.messageLength, 3); err != nil {
					return err
				}
				if err := utils.WriteByte(w, cs.messageTypeID); err != nil {
					return err
				}
				if err := utils.WriteUintLE(w, cs.messageStreamID, 4); err != nil {
					return err
				}
				if err := utils.WriteUintBE(w, cs.clock, 4); err != nil {
					return err
				}
			} else {
				if err := utils.WriteUintBE(w, cs.clock, 3); err != nil {
					return err
				}
				if err := utils.WriteUintBE(w, cs.messageLength, 3); err != nil {
					return err
				}
				if err := utils.WriteByte(w, cs.messageTypeID); err != nil {
					return err
				}
				if err := utils.WriteUintLE(w, cs.messageStreamID, 4); err != nil {
					return err
				}
			}
		}
		//data
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end > cs.messageLength {
			end = cs.messageLength
		}
		if err := utils.WriteBytes(w, cs.data[start:end]); err != nil {
			return err
		}
	}
	return nil
}

func (cs *chunkStream) writeToData(v []byte) error {
	if uint32(len(v))+cs.dataIndex > cs.messageLength {
		return fmt.Errorf("write too much data to chunk stream, message length: %d, data length: %d, data index: %d",
			cs.messageLength, len(v), cs.dataIndex)
	}
	copy(cs.data[cs.dataIndex:uint32(len(v))+cs.dataIndex], v)
	cs.dataIndex += uint32(len(v))
	return nil
}
