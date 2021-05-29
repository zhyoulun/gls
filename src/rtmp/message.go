package rtmp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/zhyoulun/gls/src/av"
	"github.com/zhyoulun/gls/src/core"
	"github.com/zhyoulun/gls/src/utils"
)

type message struct {
	cs *chunkStream
}

func newMessage(cs *chunkStream) (*message, error) {
	return &message{
		cs: cs,
	}, nil
}

func (m *message) toCsvHeader() string {
	return fmt.Sprintf("chunkStreamID,timestamp,messageLength,messageTypeID,messageStreamID\n")
}

func (m *message) toCsvLine() string {
	return fmt.Sprintf("%d,%d,%d,%d,%d\n", m.cs.chunkStreamID, m.cs.clock,
		m.cs.messageLength, m.cs.messageTypeID, m.cs.messageStreamID)
}

func (m *message) getChunkStreamID() uint32 {
	return m.cs.chunkStreamID
}

func (m *message) GetMessageStreamID() uint32 {
	return m.cs.messageStreamID
}

func (m *message) getMessageTypeID() uint8 {
	return m.cs.messageTypeID
}

func (m *message) GetData() []byte {
	return m.cs.data
}

func (m *message) GetTimestamp() uint32 {
	return m.cs.clock
}

func (m *message) GetAvType() (uint8, error) {
	var avType uint8
	if m.cs.messageTypeID == typeAudio {
		avType = av.TypeAudio
	} else if m.cs.messageTypeID == typeVideo {
		avType = av.TypeVideo
	} else if m.cs.messageTypeID == typeDataAMF0 || m.cs.messageTypeID == typeDataAMF3 {
		avType = av.TypeMetadata
	} else {
		return 0, core.ErrorImpossible
	}
	return avType, nil
}

func newBaseProtocolControlMessage(messageLength uint32, messageTypeID uint8) (*chunkStream, error) {
	var timestamp uint32 = 0 //ignored
	return newChunkStreamForMessage(chunkStreamID2, timestamp, messageLength, messageTypeID, messageStreamID0)
}

//protocol control message
func newPCMWindowAcknowledgementSize(size uint32) (*chunkStream, error) {
	cs, err := newBaseProtocolControlMessage(4, typeWindowAcknowledgementSize)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	if err := cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return cs, nil
}

//protocol control message
func newPCMSetPeerBandwidth(size uint32) (*chunkStream, error) {
	cs, err := newBaseProtocolControlMessage(5, typeSetPeerBandwidth)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(limitTypeDynamic); err != nil {
		return nil, err
	}
	if err := cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return cs, nil
}

//protocol control message
func newPCMSetChunkSize(size uint32) (*chunkStream, error) {
	cs, err := newBaseProtocolControlMessage(4, typeSetChunkSize)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	if err := cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return cs, nil
}

func newCommandMessage(chunkStreamID, messageStreamID uint32, data []byte) (*chunkStream, error) {
	cs, err := newChunkStreamForMessage(chunkStreamID, 0, uint32(len(data)), typeCommandAMF0, messageStreamID) //todo 为什么没有amf3
	if err != nil {
		return nil, err
	}
	if err := cs.writeToData(data); err != nil {
		return nil, err
	}
	return cs, nil
}

func newBaseUserControlMessage(eventType, dataLength uint32) (*chunkStream, error) {
	cs, err := newChunkStreamForMessage(2, 0, dataLength+2, typeUserControl, 1)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := utils.WriteUintBE(buf, eventType, 2); err != nil {
		return nil, err
	}
	if err := cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return cs, nil
}

func newUCMStreamIsRecorded(messageStreamID uint32) (*chunkStream, error) {
	cs, err := newBaseUserControlMessage(EventStreamIsRecorded, 4)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := utils.WriteUintBE(buf, messageStreamID, 4); err != nil {
		return nil, err
	}
	if err := cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return cs, nil
}

func newUCMStreamBegin(messageStreamID uint32) (*chunkStream, error) {
	cs, err := newBaseUserControlMessage(EventStreamBegin, 4)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := utils.WriteUintBE(buf, messageStreamID, 4); err != nil {
		return nil, err
	}
	if err := cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return cs, nil
}
