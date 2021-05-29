package rtmp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/zhyoulun/gls/src/av"
	"github.com/zhyoulun/gls/src/core"
	"github.com/zhyoulun/gls/src/flv"
	"github.com/zhyoulun/gls/src/utils"
)

type message struct {
	cs *chunkStream
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

func (m *message) String() string {
	return fmt.Sprintf("timestamp: %d, messageLength: %d, messageTypeID: %d, messageStreamID: %d, dataIndex: %d",
		m.cs.clock, m.cs.messageLength, m.cs.messageTypeID, m.cs.messageStreamID, m.cs.dataIndex)
}

func newMessage(cs *chunkStream) (*message, error) {
	return &message{
		cs: cs,
	}, nil
}

func newMessage2(chunkStreamID, timestamp, messageLength uint32, messageTypeID uint8, messageStreamID uint32) (*message, error) {
	cs := &chunkStream{
		chunkStreamID:   chunkStreamID,
		clock:           timestamp,
		messageLength:   messageLength,
		messageTypeID:   messageTypeID,
		messageStreamID: messageStreamID,
		data:            make([]byte, messageLength),
		dataIndex:       0,
	}
	return &message{cs: cs}, nil
}

func newMessage3(p *av.Packet) (*message, error) {
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

	cs := &chunkStream{
		chunkStreamID:   chunkStreamID,
		clock:           p.GetTimestamp(),
		messageLength:   dataLength,
		messageTypeID:   messageTypeID,
		messageStreamID: p.GetStreamID(),
		data:            data,
		dataIndex:       0,
	}
	return &message{cs: cs}, nil
}

func newBaseProtocolControlMessage(messageLength uint32, messageTypeID uint8) (*message, error) {
	var timestamp uint32 = 0 //ignored
	return newMessage2(chunkStreamID2, timestamp, messageLength, messageTypeID, messageStreamID0)
}

func newPCMAcknowledgement(size uint32) (*message, error) {
	m, err := newBaseProtocolControlMessage(4, typeAcknowledgement)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	if err := m.cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return m, nil
}

//protocol control message
func newPCMWindowAcknowledgementSize(size uint32) (*message, error) {
	m, err := newBaseProtocolControlMessage(4, typeWindowAcknowledgementSize)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	if err := m.cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return m, nil
}

//protocol control message
func newPCMSetPeerBandwidth(size uint32) (*message, error) {
	m, err := newBaseProtocolControlMessage(5, typeSetPeerBandwidth)
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
	if err := m.cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return m, nil
}

//protocol control message
func newPCMSetChunkSize(size uint32) (*message, error) {
	m, err := newBaseProtocolControlMessage(4, typeSetChunkSize)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	if err := m.cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return m, nil
}

func newCommandMessage(chunkStreamID, messageStreamID uint32, data []byte) (*message, error) {
	m, err := newMessage2(chunkStreamID, 0, uint32(len(data)), typeCommandAMF0, messageStreamID)
	if err != nil {
		return nil, err
	}
	if err := m.cs.writeToData(data); err != nil {
		return nil, err
	}
	return m, nil
}

func newBaseUserControlMessage(eventType, dataLength uint32) (*message, error) {
	var timestamp uint32 = 0 //ignored
	m, err := newMessage2(chunkStreamID2, timestamp, dataLength+2, typeUserControl, messageStreamID0)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := utils.WriteUintBE(buf, eventType, 2); err != nil {
		return nil, err
	}
	if err := m.cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return m, nil
}

func newUCMStreamIsRecorded(messageStreamID uint32) (*message, error) {
	m, err := newBaseUserControlMessage(EventStreamIsRecorded, 4)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := utils.WriteUintBE(buf, messageStreamID, 4); err != nil {
		return nil, err
	}
	if err := m.cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return m, nil
}

func newUCMStreamBegin(messageStreamID uint32) (*message, error) {
	m, err := newBaseUserControlMessage(EventStreamBegin, 4)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := utils.WriteUintBE(buf, messageStreamID, 4); err != nil {
		return nil, err
	}
	if err := m.cs.writeToData(buf.Bytes()); err != nil {
		return nil, err
	}
	return m, nil
}
