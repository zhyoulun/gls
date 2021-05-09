package rtmp

import (
	"bytes"
	"encoding/binary"
	"github.com/zhyoulun/gls/src/utils"
)

func newBaseProtocolControlMessage(messageLength uint32, messageTypeID uint8) (*chunkStream, error) {
	return newChunkStream2(2, 0, messageLength, messageTypeID, 0)
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
	if err := cs.WriteToData(buf.Bytes()); err != nil {
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
	if err := cs.WriteToData(buf.Bytes()); err != nil {
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
	if err := cs.WriteToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return cs, nil
}

func newCommandMessage(chunkStreamID, messageStreamID uint32, data []byte) (*chunkStream, error) {
	cs, err := newChunkStream2(chunkStreamID, 0, uint32(len(data)), typeCommandAMF0, messageStreamID) //todo 为什么没有amf3
	if err != nil {
		return nil, err
	}
	if err := cs.WriteToData(data); err != nil {
		return nil, err
	}
	return cs, nil
}

func newBaseUserControlMessage(eventType, dataLength uint32) (*chunkStream, error) {
	cs, err := newChunkStream2(2, 0, dataLength+2, typeUserControl, 1)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := utils.WriteUintBE(buf, eventType, 2); err != nil {
		return nil, err
	}
	if err := cs.WriteToData(buf.Bytes()); err != nil {
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
	if err := cs.WriteToData(buf.Bytes()); err != nil {
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
	if err := cs.WriteToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return cs, nil
}
