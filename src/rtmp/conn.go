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

	localWindowAckSize  uint32
	remoteWindowAckSize uint32
	receivedSize        uint32

	localPeerBandwidth uint32

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

		localWindowAckSize:  2.5e6, //todo ??
		remoteWindowAckSize: 2.5e6, //todo ??

		localPeerBandwidth: 2.5e6, //todo ??

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
	case typeSharedObjectAMF0: //todo

	case typeCommandAMF3: //todo
		return errors.Wrapf(core.ErrorNotImplemented, "typeCommandAMF3")
	case typeCommandAMF0:
		return c.handleCommandMessage(chunkStreamID, messageStreamID, messageTypeID, data, timestamp)
	case typeAggregate: //todo
	default:
		return errors.Wrapf(core.ErrorNotSupported, "messageTypeID: %d", messageTypeID)
	}

	return nil
}

func (c *Conn) handleCommandMessage(chunkStreamID, messageStreamID uint32, typeID uint8, data []byte, timestamp uint32) error {
	log.Tracef("handle command message, chunkStreamID: %d, messageStreamID: %d, typeID: %d", chunkStreamID, messageStreamID, typeID)
	amfDecoder, err := amf.NewAmf()
	if err != nil {
		return err
	}
	r := bytes.NewReader(data)
	vs, err := amfDecoder.DecodeBatch(r, amf.Amf0)
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
	case commandNetConnectionConnect:
		return c.handleCommandConnect(chunkStreamID, messageStreamID, vs[1:])
	case commandNetConnectionCall: //todo
	case commandNetConnectionCreateStream:
		return c.handleCommandCreateStream(chunkStreamID, messageStreamID, vs[1:])
	case commandNetStreamPlay:
		c.readMessageDone = true
		return c.handleCommandPlay(chunkStreamID, messageStreamID, vs[1:])
	case commandNetStreamPlay2: //todo
	case commandNetStreamDeleteStream: //todo
	case commandNetStreamCloseStream: //todo
	case commandNetStreamReceiveAudio: //todo
	//NetStream send the receiveAudio message to inform the server whether to send or not to send the audio to the client
	case commandNetStreamReceiveVideo: //todo
	case commandNetStreamPublish:
		c.readMessageDone = true
		c.isPublish = true
		return c.handleCommandPublish(chunkStreamID, messageStreamID, vs[1:])
	case commandNetStreamSeek: //todo
	case commandNetStreamPause: //todo
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
		if n != 5 {
			return fmt.Errorf("set peer bandwidth error, length: %d", n)
		}
		//todo peer bandwidth, limit type strategy
	default:
		return core.ErrorNotSupported
	}
	return nil
}

//1. client -> server: send connect command(connect)
//2. server -> client: window acknowledgement size
//3. server -> client: set peer bandwidth
//4. client -> server: window acknowledgement size -- 这个不在该函数中处理
//5. server -> client: user control message(StreamBegin)
//6. server -> client: command message(_result - connect response)
func (c *Conn) handleCommandConnect(chunkStreamID, messageStreamID uint32, vs []interface{}) error {
	if len(vs) != 2 {
		return errors.Errorf("handle command connect, len(vs) error, want: 2, got: %d", len(vs))
	}

	var transactionID float64
	var ok bool
	transactionID, ok = vs[0].(float64)
	if !ok {
		return fmt.Errorf("parse transaction id fail, not float64")
	}
	if transactionID != transactionID1 {
		return fmt.Errorf("parse transcation id, want: 1, got: %f", transactionID)
	}

	obj, ok := vs[1].(amf.AmfObject)
	if !ok {
		return fmt.Errorf("parse Command Object fail, not AmfObject")
	}
	log.Debugf("command object: %s", obj)
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

	log.Infof("handle command connect, connInfo: %+v", c.connInfo)

	if m, err := newPCMWindowAcknowledgementSize(c.localWindowAckSize); err != nil {
		return err
	} else {
		if err := c.writeMessage(m); err != nil {
			return err
		}
	}
	if cs, err := newPCMSetPeerBandwidth(c.localPeerBandwidth); err != nil {
		return err
	} else {
		if err := c.writeMessage(cs); err != nil {
			return err
		}
	}
	//todo ?? 这里和协议对不上，但livego等都在connect是实现了set chunk size，并且如果缺少这句话，无法推流
	if m, err := newPCMSetChunkSize(c.localMaximumChunkSize); err != nil {
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
	if msg, err := newNetConnectionResponseConnect(transactionID); err != nil {
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

	if m, err := newUCMStreamBegin(messageStreamID); err != nil {
		return err
	} else {
		if err := c.writeMessage(m); err != nil {
			return err
		}
	}

	if msg, err := newNetConnectionResponseCreateStream(transactionID, 1); err != nil { //todo ??这里为什么是1
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

//1. client -> server: command message(publish)
//2. server -> client: user control(stream begin)
//3. client -> server: data message(metadata)
//4. client -> server: audio data
//5. client -> server: set chunk size
//5. server -> client: command message(_result - publish result)
func (c *Conn) handleCommandPublish(chunkStreamID, messageStreamID uint32, vs []interface{}) error {
	if len(vs) < 3 {
		return errors.Errorf("handle command publish, len(vs) error, want: >=3, got: %d", len(vs))
	}
	transactionID, ok := vs[0].(float64)
	if !ok || transactionID != 5 { //todo ??
		return errors.Errorf("parse transaction id, want float64 0, got: %+v", vs[0])
	}
	//vs[1] is nil
	publishingName, ok := vs[2].(string)
	if !ok {
		return errors.Errorf("parse publishing name, want string, got: %+v", vs[2])
	}
	c.streamName = publishingName
	//vs[3] publishing type

	if msg, err := newNetStreamResponsePublishStart(); err != nil {
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

//CreateStream
//1. client -> server: command message(createStream)
//2. server -> client: command message(_result - createStream response)
//Play
//1. client -> server: command message(play)
//2. server -> client: set chunk size
//3. server -> client: user control message(stream is recorded)
//4. server -> client: user control message(stream begin)
//5. server -> client: command message(onStatus-play reset)
//6. server -> client: command message(onStatus-play start)
func (c *Conn) handleCommandPlay(chunkStreamID, messageStreamID uint32, vs []interface{}) error {
	if len(vs) < 3 {
		return errors.Errorf("handle command play, len(vs) error, want: >=3, got: %d", len(vs))
	}
	transactionID, ok := vs[0].(float64)
	if !ok || transactionID != 4 { //todo ??
		return errors.Errorf("parse transaction id, want float64 4, got: %+v", vs[0])
	}
	//vs[1] is nil
	streamName, ok := vs[2].(string)
	if !ok {
		return errors.Errorf("parse stream name, want string, got: %+v", vs[2])
	}
	c.streamName = streamName
	//vs[3]  start,number,optional
	//vs[4]  duration,number,optional
	//vs[5]  reset,number,optional

	if m, err := newPCMSetChunkSize(c.localMaximumChunkSize); err != nil {
		return err
	} else {
		if err := c.writeMessage(m); err != nil {
			return err
		}
	}

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

	if msg, err := newNetStreamResponsePlayReset(); err != nil {
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

	if msg, err := newNetStreamResponsePlayStart(); err != nil {
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
