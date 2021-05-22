package rtmp

import (
	"encoding/binary"
	"fmt"
	"github.com/zhyoulun/gls/src/av"
	"github.com/zhyoulun/gls/src/core"
	"github.com/zhyoulun/gls/src/flv"
	"github.com/zhyoulun/gls/src/utils"
	"github.com/zhyoulun/gls/src/utils/debug"
	"io"
)

type chunkStream struct {
	fmt             Fmt
	chunkStreamID   uint32
	timestamp       uint32
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
		cs.chunkStreamID, cs.timestamp, cs.messageLength, cs.messageTypeID, cs.messageStreamID,
		cs.tmp.extended, cs.tmp.timestampDelta, cs.dataIndex)
}

func (cs *chunkStream) toCsvHeader() string {
	return fmt.Sprintf("chunkStreamID,timestamp,messageLength,messageTypeID,messageStreamID\n")
}

func (cs *chunkStream) toCsvLine() string {
	return fmt.Sprintf("%d,%d,%d,%d,%d\n", cs.chunkStreamID, cs.timestamp,
		cs.messageLength, cs.messageTypeID, cs.messageStreamID)
}

func newChunkStreamForRead(basicHeader *chunkBasicHeader) (*chunkStream, error) {
	return &chunkStream{
		chunkStreamID: basicHeader.chunkStreamID,
		tmp: chunkStreamTmp{
			currentFmt: basicHeader.fmt,
		},
	}, nil
}

func newChunkStreamForMessage(chunkStreamID, timestamp, messageLength uint32, messageTypeID uint8, messageStreamID uint32) (*chunkStream, error) {
	return &chunkStream{
		chunkStreamID:   chunkStreamID,
		timestamp:       timestamp,
		messageLength:   messageLength,
		messageTypeID:   messageTypeID,
		messageStreamID: messageStreamID,
		data:            make([]byte, messageLength),
		dataIndex:       0,
	}, nil
}

func newChunkStreamForPacket(p *av.Packet) (*chunkStream, error) {
	var messageTypeID uint8
	avType := p.GetAvType()
	if avType == av.TypeAudio {
		messageTypeID = typeAudio
	} else if avType == av.TypeVideo {
		messageTypeID = typeVideo
	} else if avType == av.TypeMetadata {
		messageTypeID = typeDataAMF0
	} else {
		return nil, core.ErrorImpossible
	}

	//获取data和dataLength
	data := p.GetData()
	var err error
	if messageTypeID == typeDataAMF0 {
		if data, err = flv.MetadataReformDelete(data); err != nil {
			return nil, err
		}
	}
	dataLength := uint32(len(data))

	var chunkStreamID uint32
	if avType == av.TypeAudio {
		chunkStreamID = 4 //todo ??
	} else if avType == av.TypeVideo || avType == av.TypeMetadata {
		chunkStreamID = 6 //todo ??
	} else {
		return nil, core.ErrorImpossible
	}

	return &chunkStream{
		chunkStreamID:   chunkStreamID,
		timestamp:       p.GetTimestamp(),
		messageLength:   dataLength,
		messageTypeID:   messageTypeID,
		messageStreamID: p.GetStreamID(),
		data:            data,
		dataIndex:       0,
	}, nil
}

//chunk stream中的临时变量，对chunk stream本身没作用，主要是用来基于若干个chunk组合出一个完整的message based on chunk stream
type chunkStreamTmp struct {
	currentFmt         Fmt  //chunk type
	extended           bool //是否需要读取extend timestamp，用于fmt3 chunk的读取场景
	firstChunkReadDone bool //chunkStream中，第一个chunk读取完成
	timestampDelta     uint32
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
		cs.timestamp, cs.messageLength, cs.messageTypeID, cs.messageStreamID, cs.dataIndex)
}

func (cs *chunkStream) got() bool {
	return cs.messageLength == cs.dataIndex
}

func (cs *chunkStream) readChunk(r utils.ReadPeeker, chunkSize uint32) error {
	switch cs.tmp.currentFmt {
	case fmt0: //11B, this type must be used at the start of a chunk stream
		cs.fmt = fmt0
		var timestamp uint32
		var err error
		if timestamp, err = utils.ReadUintBE(r, 3); err != nil {
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
		if cs.messageStreamID, err = utils.ReadUintLE(r, 4); err != nil {
			return err
		}
		//cs.timestamp
		if timestamp == 0xffffff {
			var extendedTimestamp uint32
			if extendedTimestamp, err = utils.ReadUintBE(r, 4); err != nil {
				return err
			}
			cs.timestamp = extendedTimestamp
			cs.tmp.extended = true
			cs.tmp.timestampDelta = 0
		} else {
			cs.timestamp = timestamp
			cs.tmp.extended = false
			cs.tmp.timestampDelta = 0
		}
		//cs.data init
		cs.data = make([]byte, cs.messageLength)
		cs.dataIndex = 0
		cs.tmp.firstChunkReadDone = true
	case fmt1: //7B, stream with variable-sized message(for example: many video formats) should use this format for the first chunk of each new message after the first
		cs.fmt = fmt1
		var err error
		var timestampDelta uint32
		if timestampDelta, err = utils.ReadUintBE(r, 3); err != nil {
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
		if timestampDelta == 0xffffff {
			var extendedTimestamp uint32
			if extendedTimestamp, err = utils.ReadUintBE(r, 4); err != nil {
				return err
			}
			cs.timestamp += extendedTimestamp
			cs.tmp.extended = true
			cs.tmp.timestampDelta = extendedTimestamp
		} else {
			cs.timestamp += timestampDelta
			cs.tmp.extended = false
			cs.tmp.timestampDelta = timestampDelta
		}
		//cs.data init
		cs.data = make([]byte, cs.messageLength)
		cs.dataIndex = 0
		cs.tmp.firstChunkReadDone = true
	case fmt2: //3B, stream with constant-sized messages(for example: some audio and data formats) should use this format for the first chunk of each message after the first
		cs.fmt = fmt2
		var err error
		var timestampDelta uint32
		if timestampDelta, err = utils.ReadUintBE(r, 3); err != nil {
			return err
		}
		//cs.timestamp
		if timestampDelta == 0xffffff {
			var extendedTimestamp uint32
			if extendedTimestamp, err = utils.ReadUintBE(r, 4); err != nil {
				return err
			}
			cs.timestamp += extendedTimestamp
			cs.tmp.extended = true
			cs.tmp.timestampDelta = extendedTimestamp
		} else {
			cs.timestamp += timestampDelta
			cs.tmp.extended = false
			cs.tmp.timestampDelta = timestampDelta
		}
		//cs.data init
		cs.data = make([]byte, cs.messageLength)
		cs.dataIndex = 0
		cs.tmp.firstChunkReadDone = true
	case fmt3: //0B; fmt3不可能是ChunkStream的第一个chunk
		var err error
		if cs.dataIndex == cs.messageLength { //todo ??不明白
			//cs.timestamp
			switch cs.fmt {
			case fmt0:
				if cs.tmp.extended {
					var extendedTimestamp uint32
					if extendedTimestamp, err = utils.ReadUintBE(r, 4); err != nil {
						return err
					}
					cs.timestamp = extendedTimestamp
				}
				//todo 为什么这里没有else？
			case fmt1, fmt2:
				var timestampDelta uint32
				if cs.tmp.extended {
					var extendedTimestamp uint32
					if extendedTimestamp, err = utils.ReadUintBE(r, 4); err != nil {
						return err
					}
					timestampDelta = extendedTimestamp
				} else {
					timestampDelta = cs.tmp.timestampDelta
				}
				cs.timestamp += timestampDelta
			}
			//cs.data init
			cs.data = make([]byte, cs.messageLength) //todo 这里的cs.messageLength从哪里来？从上个有相同chunkStreamID的message来？
			cs.dataIndex = 0
			cs.tmp.firstChunkReadDone = true
		} else {
			if cs.tmp.extended {
				//todo 这段逻辑比较神奇
				b, err := r.Peek(4)
				if err != nil {
					return err
				}
				tmpTS := binary.BigEndian.Uint32(b)
				if tmpTS == cs.timestamp {
					_, _ = utils.ReadBytes(r, 4) //discard
				}
			}
			//todo 为啥没有else呢？
		}
	default:
		return core.ErrorImpossible
	}

	//cs.data read
	//一般读chunkSize个字节数，最后一次可能比chunkSize小
	readLength := cs.messageLength - cs.dataIndex
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
			if cs.timestamp > 0xffffff {
				//todo ??为什么要删掉
				//if err := utils.WriteUintBE(w, 0xffffff, 3); err != nil {
				//	return err
				//}
				if err := utils.WriteUintBE(w, cs.timestamp, 4); err != nil {
					return err
				}
			} else {
				//todo ??为什么要删掉
				//if err := utils.WriteUintBE(w, cs.timestamp, 3); err != nil {
				//	return err
				//}
			}
		} else if f == fmt0 {
			if cs.timestamp > 0xffffff {
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
				if err := utils.WriteUintBE(w, cs.timestamp, 4); err != nil {
					return err
				}
			} else {
				if err := utils.WriteUintBE(w, cs.timestamp, 3); err != nil {
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

func (cs *chunkStream) GetAvType() (uint8, error) {
	var avType uint8
	if cs.messageTypeID == typeAudio {
		avType = av.TypeAudio
	} else if cs.messageTypeID == typeVideo {
		avType = av.TypeVideo
	} else if cs.messageTypeID == typeDataAMF0 || cs.messageTypeID == typeDataAMF3 {
		avType = av.TypeMetadata
	} else {
		return 0, core.ErrorImpossible
	}
	return avType, nil
}

func (cs *chunkStream) GetMessageStreamID() uint32 {
	return cs.messageStreamID
}

func (cs *chunkStream) GetTimestamp() uint32 {
	return cs.timestamp
}

func (cs *chunkStream) GetData() []byte {
	return cs.data
}
