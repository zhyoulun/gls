package rtmp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/zhyoulun/gls/src/amf"
	"github.com/zhyoulun/gls/src/av"
	"github.com/zhyoulun/gls/src/core"
	"github.com/zhyoulun/gls/src/flv"
	"github.com/zhyoulun/gls/src/utils"
	"github.com/zhyoulun/gls/src/utils/debug"
)

type Conn struct {
	conn utils.PeekerConn

	localMaximumChunkSize  uint32
	remoteMaximumChunkSize uint32 //the maximum chunk size should be at least 128 bytes, and must be at least 1 byte
	remoteWindowAckSize    uint32
	receivedSize           uint32

	chunkStreams map[uint32]*chunkStream

	readMessageDone bool //默认值为false
	isPublish       bool //默认值为false //todo

	connInfo   ConnectCommentObject
	streamName string
}

func NewConn(conn utils.PeekerConn) (*Conn, error) {
	return &Conn{
		conn: conn,

		localMaximumChunkSize:  defaultLocalMaximumChunkSize,
		remoteMaximumChunkSize: defaultRemoteMaximumChunkSize,
		remoteWindowAckSize:    2500000, //todo ??

		chunkStreams: make(map[uint32]*chunkStream),
	}, nil
}

func (c *Conn) Handshake() error {
	var h *handshake
	var err error
	if h, err = newHandshake(); err != nil {
		return err
	}
	if err = h.Do(c.conn); err != nil {
		return err
	}
	return nil
}

func (c *Conn) ReadHeader() error {
	for {
		if c.readMessageDone {
			break
		}
		var m *message
		var err error
		if m, err = c.readMessage(); err != nil {
			return err
		}
		//handle message in chunk stream
		if err = c.handleMessage(m.getChunkStreamID(), m.GetMessageStreamID(), m.getMessageTypeID(), m.GetData(), m.GetTimestamp()); err != nil { //todo 这里的cs.timestamp传参可能有问题
			return err
		}
	}
	return nil
}

func (c *Conn) IsPublish() bool {
	return c.isPublish
}

func (c *Conn) GetStreamName() string {
	return c.streamName
}

func (c *Conn) GetConnInfo() ConnectCommentObject {
	return c.connInfo
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) ReadPacket() (*av.Packet, error) {
	var m *message
	var err error
	for {
		m, err = c.readMessage()
		if err != nil {
			return nil, err
		}

		//debug
		debug.Csv.Write(&debug.Message{
			FileName:   "message.csv",
			HeaderLine: m.toCsvHeader(),
			BodyLine:   m.toCsvLine(),
		})

		if m.getMessageTypeID() == typeAudio || m.getMessageTypeID() == typeVideo ||
			m.getMessageTypeID() == typeDataAMF0 || m.getMessageTypeID() == typeDataAMF3 {
			break
		}
		log.Tracef("read packet, ignore, messageTypeID: %d", m.getMessageTypeID())
	}

	demuxer := flv.NewDemuxer()
	p, err := av.NewPacket(m, demuxer)
	if err != nil {
		return nil, err
	}
	//debug
	debug.Csv.Write(&debug.Message{
		FileName:   "packet.csv",
		HeaderLine: p.ToCsvHeader(),
		BodyLine:   p.ToCsvLine(),
	})
	return p, nil
}

func (c *Conn) WritePacket(p *av.Packet) error {
	m, err := newMessage3(p)
	if err != nil {
		return err
	}
	if err := c.writeMessage(m); err != nil {
		return err
	}
	log.Tracef("write message: %s", m)
	return nil
}

func (c *Conn) readMessage() (*message, error) {
	var cs *chunkStream
	for {
		var basicHeader *chunkBasicHeader
		var err error
		var ok bool

		//read chunk basic header
		if basicHeader, err = newChunkBasicHeaderForRead(); err != nil {
			return nil, err
		}
		if err = basicHeader.Read(c.conn); err != nil {
			return nil, err
		}
		log.Tracef("basic header: %s", basicHeader)

		//init chunk stream
		if cs, ok = c.chunkStreams[basicHeader.chunkStreamID]; !ok {
			if cs, err = newChunkStreamForRead(basicHeader); err != nil {
				return nil, err
			}
			c.chunkStreams[basicHeader.chunkStreamID] = cs
			log.Infof("got a new chunk stream, chunkStreamID: %d", basicHeader.chunkStreamID)
		} else {
			cs.setBasicHeader(basicHeader) //这里容易遗漏！
		}

		//read chunk to chunk stream
		if err = cs.readChunk(c.conn, c.remoteMaximumChunkSize); err != nil {
			return nil, err
		}
		if cs.gotOneMessage() {
			log.Tracef("got chunk stream: %s", cs)
			break
		}
	}

	if err := c.ack(cs.messageLength); err != nil {
		return nil, err
	}

	return newMessage(cs)
}

func (c *Conn) ack(receivedSize uint32) error {
	c.receivedSize += receivedSize
	if c.receivedSize >= c.remoteWindowAckSize {
		m, err := newPCMAcknowledgement(receivedSize)
		if err != nil {
			return err
		}
		if err := c.writeMessage(m); err != nil {
			return err
		}
		c.receivedSize = 0
	}
	return nil
}

func (c *Conn) writeMessage(m *message) error {
	return m.cs.writeChunk(c.conn, c.localMaximumChunkSize)
}

func (c *Conn) handleMessage(chunkStreamID, messageStreamID uint32, messageTypeID uint8, data []byte, timestamp uint32) error {
	log.Tracef("rtmp conn handle message start, chunkStreamID: %d, messageStreamID: %d, messageTypeID: %d, timestamp: %d", chunkStreamID, messageStreamID, messageTypeID, timestamp)
	defer func() {
		log.Tracef("rtmp conn handle message end, chunkStreamID: %d, messageStreamID: %d, messageTypeID: %d, timestamp: %d", chunkStreamID, messageStreamID, messageTypeID, timestamp)
	}()
	switch messageTypeID {
	case typeSetChunkSize,
		typeAbort,
		typeAcknowledgement,
		typeWindowAcknowledgementSize,
		typeSetPeerBandwidth: //timestamp is ignored
		if chunkStreamID != chunkStreamID2 {
			return errors.Errorf("protocol control message must be sent in chunk stream id 2")
		}
		if messageStreamID != messageStreamID0 {
			return errors.Errorf("protocol control message must have message stream id 0")
		}
		return c.handleProtocolControlMessage(messageTypeID, data)
	case typeUserControl: //todo

	case typeAudio: //todo
	case typeVideo: //todo

	case typeDataAMF3: //todo
	case typeDataAMF0: //todo

	case typeSharedObjectAMF3: //todo
		return nil
	case typeSharedObjectAMF0: //todo
		return nil

	case typeCommandAMF3,
		typeCommandAMF0:
		return c.handleCommandMessage(chunkStreamID, messageStreamID, messageTypeID, data, timestamp)
	case typeAggregate: //todo
		return nil
	default:
		return errors.Wrapf(core.ErrorNotSupported, "messageTypeID: %d", messageTypeID)
	}

	return nil
}

func (c *Conn) handleCommandMessage(chunkStreamID, messageStreamID uint32, typeID uint8, data []byte, timestamp uint32) error {
	log.Tracef("handle command message, chunkStreamID: %d, messageStreamID: %d, typeID: %d", chunkStreamID, messageStreamID, typeID)
	if typeID == typeCommandAMF3 { //todo ?
		data = data[1:]
	}
	amfDecoder, err := amf.NewAmf()
	if err != nil {
		return err
	}
	r := bytes.NewReader(data)
	vs, err := amfDecoder.DecodeBatch(r, amf.Amf0) //todo 需要确认这里只有amf0吗？
	if err != nil {
		return errors.Wrap(err, "amf decoder decode batch")
	}
	if len(vs) == 0 {
		return fmt.Errorf("amf decode error, len(vs)=0")
	}

	switch vs[0].(type) {
	case string:
	default:
		return core.ErrorUnknown
	}

	command, ok := vs[0].(string)
	if !ok {
		return core.ErrorImpossible
	}

	switch command {
	case commandConnect:
		return c.handleCommandConnect(chunkStreamID, messageStreamID, vs[1:])
	case commandCall: //todo
	case commandCreateStream:
		return c.handleCommandCreateStream(chunkStreamID, messageStreamID, vs[1:])
	case commandPlay:
		c.readMessageDone = true
		return c.handleCommandPlay(chunkStreamID, messageStreamID, vs[1:])
	case commandPlay2: //todo
	case commandDeleteStream: //todo
	case commandCloseStream: //todo
	case commandReceiveAudio: //todo
	case commandReceiveVideo: //todo
	case commandPublish:
		c.readMessageDone = true
		c.isPublish = true
		return c.handleCommandPublish(chunkStreamID, messageStreamID, vs[1:])
	case commandSeek: //todo
	case commandPause: //todo
	case commandOnStatus: //todo
	default:
		log.Warnf("unknown command: %s", command)
	}
	return nil
}

func (c *Conn) handleProtocolControlMessage(typeID uint8, data []byte) error {
	n := len(data)
	switch typeID {
	case typeSetChunkSize:
		if n != 4 {
			return fmt.Errorf("set chunk size error, length: %d", n)
		}
		v := binary.BigEndian.Uint32(data)
		if v > maxValidMaximumChunkSize || v < minValidMaximumChunkSize {
			return fmt.Errorf("invalid maximum chunck size, value: %d", v)
		}
		c.remoteMaximumChunkSize = v
	case typeAbort:
		if n != 4 {
			return fmt.Errorf("abort error, length: %d", n)
		}
		//todo discard the partially received message over a chunk stream
	case typeAcknowledgement:
		if n != 4 {
			return fmt.Errorf("acknowledgement error, length: %d", n)
		}
		//check is useless
	case typeWindowAcknowledgementSize:
		if n != 4 {
			return fmt.Errorf("window acknowledgement size error, length: %d", n)
		}
		c.remoteWindowAckSize = binary.BigEndian.Uint32(data)
	case typeSetPeerBandwidth:
		//todo
		return core.ErrorNotImplemented
	default:
		return core.ErrorNotSupported
	}
	return nil
}

func (c *Conn) handleCommandConnect(chunkStreamID, messageStreamID uint32, vs []interface{}) error {
	if len(vs) != 2 {
		return errors.Errorf("handle command connect, len(vs) error, want: 2, got: %d", len(vs))
	}
	var transactionID float64
	switch vs[0].(type) {
	case float64:
		transactionID = vs[0].(float64)
		if transactionID != 1 {
			return fmt.Errorf("parse transcation id, want: 1, got: %f", transactionID)
		}
	}
	switch vs[1].(type) {
	case amf.AmfObject:
		obj := vs[1].(amf.AmfObject)
		if v, ok := obj["app"]; ok {
			c.connInfo.App = v.(string)
		}
		if v, ok := obj["flashver"]; ok {
			c.connInfo.Flashver = v.(string)
		}
		if v, ok := obj["swfUrl"]; ok {
			c.connInfo.SwfUrl = v.(string)
		}
		if v, ok := obj["tcUrl"]; ok {
			c.connInfo.TcUrl = v.(string)
		}
		if v, ok := obj["fpad"]; ok {
			c.connInfo.Fpad = v.(bool)
		}
		if v, ok := obj["audioCodecs"]; ok {
			c.connInfo.AudioCodecs = int(v.(float64))
		}
		if v, ok := obj["videoCodecs"]; ok {
			c.connInfo.VideoCodecs = int(v.(float64))
		}
		if v, ok := obj["videoFunction"]; ok {
			c.connInfo.VideoFunction = int(v.(float64))
		}
		if v, ok := obj["pageUrl"]; ok {
			c.connInfo.PageUrl = v.(string)
		}
		if v, ok := obj["objectEncoding"]; ok {
			c.connInfo.ObjectEncoding = v.(float64)
		}
		if v, ok := obj["type"]; ok {
			c.connInfo.Type = v.(string)
		}
	}
	log.Tracef("handle command connect, connInfo: %+v", c.connInfo)

	if cs, err := newPCMWindowAcknowledgementSize(2.5e6); err != nil { //todo 为什么是2.5e6
		return err
	} else {
		if err := c.writeMessage(cs); err != nil {
			return err
		}
	}
	if cs, err := newPCMSetPeerBandwidth(2.5e6); err != nil { //todo 为什么是2.5e6
		return err
	} else {
		if err := c.writeMessage(cs); err != nil {
			return err
		}
	}
	if cs, err := newPCMSetChunkSize(c.localMaximumChunkSize); err != nil { //todo 是rc.chunkSize吗？
		return err
	} else {
		if err := c.writeMessage(cs); err != nil {
			return err
		}
	}
	if msg, err := newNetConnectionConnectResp(transactionID, c.connInfo.ObjectEncoding); err != nil {
		return err
	} else {
		if m, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := c.writeMessage(m); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Conn) handleCommandCreateStream(chunkStreamID, messageStreamID uint32, vs []interface{}) error {
	if len(vs) != 2 {
		return errors.Errorf("handle command connect, len(vs) error, want: 2, got: %d", len(vs))
	}
	var transactionID float64
	switch vs[0].(type) {
	case float64:
		transactionID = vs[0].(float64)
		if transactionID != 4 && transactionID != 2 { //todo ??? 4 for publish, 2 for play
			return errors.Errorf("parse transcation id, want: 4/2, got: %f", transactionID)
		}
	}

	if msg, err := newCreateStreamResp(transactionID, 1); err != nil { //todo ??这里为什么是1
		return err
	} else {
		if m, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := c.writeMessage(m); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Conn) handleCommandPublish(chunkStreamID, messageStreamID uint32, vs []interface{}) error {
	if len(vs) < 3 {
		return errors.Errorf("handle command publish, len(vs) error, want: >=3, got: %d", len(vs))
	}
	transactionID, ok := vs[0].(float64)
	if !ok || transactionID != 5 { //todo ??
		return errors.Errorf("parse transaction id, want float64 0, got: %+v", vs[0])
	}
	publishingName, ok := vs[2].(string)
	if !ok {
		return errors.Errorf("parse publishing name, want string, got: %+v", vs[2])
	}
	c.streamName = publishingName

	if msg, err := newNetStreamPublishResp(); err != nil {
		return err
	} else {
		if m, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := c.writeMessage(m); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Conn) handleCommandPlay(chunkStreamID, messageStreamID uint32, vs []interface{}) error {
	if len(vs) < 3 {
		return errors.Errorf("handle command play, len(vs) error, want: >=3, got: %d", len(vs))
	}
	transactionID, ok := vs[0].(float64)
	if !ok || transactionID != 4 { //todo ??
		return errors.Errorf("parse transaction id, want float64 4, got: %+v", vs[0])
	}
	streamName, ok := vs[2].(string)
	if !ok {
		return errors.Errorf("parse stream name, want string, got: %+v", vs[2])
	}
	c.streamName = streamName

	if m, err := newUCMStreamIsRecorded(messageStreamID); err != nil {
		return err
	} else {
		if err := c.writeMessage(m); err != nil {
			return err
		}
	}

	if m, err := newUCMStreamBegin(messageStreamID); err != nil {
		return err
	} else {
		if err := c.writeMessage(m); err != nil {
			return err
		}
	}

	if msg, err := newNetStreamPlayReset(); err != nil {
		return err
	} else {
		if m, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := c.writeMessage(m); err != nil {
				return err
			}
		}
	}

	if msg, err := newNetStreamPlayStart(); err != nil {
		return err
	} else {
		if m, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := c.writeMessage(m); err != nil {
				return err
			}
		}
	}

	//todo ??为什么需要这个
	if msg, err := newNetStreamDataStart(); err != nil {
		return err
	} else {
		if m, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := c.writeMessage(m); err != nil {
				return err
			}
		}
	}

	//todo ??为什么需要这个
	if msg, err := newNetStreamPlayPublishNotify(); err != nil {
		return err
	} else {
		if m, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := c.writeMessage(m); err != nil {
				return err
			}
		}
	}

	return nil
}
