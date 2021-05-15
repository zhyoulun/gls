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
)

type Conn struct {
	conn utils.PeekerConn

	chunkSize           uint32
	remoteChunkSize     uint32
	remoteWindowAckSize uint32
	chunkStreams        map[uint32]*chunkStream

	readMessageDone bool
	isPublish       bool //todo

	connInfo   ConnectCommentObject
	streamName string
}

func NewConn(conn utils.PeekerConn) (*Conn, error) {
	return &Conn{
		conn: conn,

		chunkSize:       1024, //defaultMaximumChunkSize,//todo ??
		remoteChunkSize: defaultMaximumChunkSize,
		chunkStreams:    make(map[uint32]*chunkStream),

		readMessageDone: false,
	}, nil
}

func (rc *Conn) Handshake() error {
	h, err := newHandshake()
	if err != nil {
		return err
	}
	if err := h.readC0(rc.conn); err != nil {
		return err
	}
	if err := h.writeS0(rc.conn); err != nil {
		return err
	}
	if err := h.writeS1(rc.conn); err != nil {
		return err
	}
	if err := h.readC1(rc.conn); err != nil {
		return err
	}
	if err := h.writeS2(rc.conn); err != nil {
		return err
	}
	if err := h.readC2(rc.conn); err != nil {
		return err
	}
	log.Tracef("handshake: %s", h)
	return nil
}

//todo 这个名字不是很好，后边尝试调整下
func (rc *Conn) ReadMessage() error {
	for {
		if rc.readMessageDone {
			break
		}
		var cs *chunkStream
		var err error
		if cs, err = rc.readChunkStream(); err != nil {
			return err
		}
		//handle message in chunk stream
		if err = rc.handleMessage(cs.chunkStreamID, cs.messageStreamID, cs.messageTypeID, cs.data, cs.timestamp); err != nil { //todo 这里的cs.timestamp传参可能有问题
			return err
		}
	}
	return nil
}

func (rc *Conn) IsPublish() bool {
	return rc.isPublish
}

func (rc *Conn) GetStreamName() string {
	return rc.streamName
}

func (rc *Conn) GetConnInfo() ConnectCommentObject {
	return rc.connInfo
}

func (rc *Conn) Close() error {
	return rc.conn.Close()
}

func (rc *Conn) ReadPacket() (*av.Packet, error) {
	var cs *chunkStream
	var err error
	for {
		cs, err = rc.readChunkStream()
		if err != nil {
			return nil, err
		}
		//todo debug
		if !utils.ChunkStreamHeaderDone {
			_, err = utils.ChunkStreamLogFile.WriteString(cs.toCsvHeader())
			if err != nil {
				log.Warnf("write log file fail: %+v", err)
			}
			utils.ChunkStreamHeaderDone = true
		}
		_, err = utils.ChunkStreamLogFile.WriteString(cs.toCsvLine())
		if err != nil {
			log.Warnf("write log file fail: %+v", err)
		}
		if cs.messageTypeID == typeAudio || cs.messageTypeID == typeVideo ||
			cs.messageTypeID == typeDataAMF0 || cs.messageTypeID == typeDataAMF3 {
			break
		}
		log.Tracef("read packet, ignore, messageTypeID: %d", cs.messageTypeID)
	}

	demuxer := flv.NewDemuxer()
	p, err := av.NewPacket(cs, demuxer)
	if err != nil {
		return nil, err
	}
	//todo debug
	if !utils.PacketHeaderDone {
		_, err = utils.PacketLogFile.WriteString(p.ToCsvHeader())
		if err != nil {
			log.Warnf("write log file fail: %+v", err)
		}
		utils.PacketHeaderDone = true
	}
	_, err = utils.PacketLogFile.WriteString(p.ToCsvLine())
	if err != nil {
		log.Warnf("write log file fail: %+v", err)
	}
	return p, nil
}

func (rc *Conn) WritePacket(p *av.Packet) error {
	cs, err := newChunkStream3(p)
	if err != nil {
		return err
	}
	if err := rc.writeChunkStream(cs); err != nil {
		return err
	}
	log.Infof("write chunk stream: %s", cs)
	return nil
}

func (rc *Conn) readChunkStream() (*chunkStream, error) {
	var cs *chunkStream
	for {
		var basicHeader *chunkBasicHeader
		var err error
		var ok bool
		if basicHeader, err = newChunkBasicHeader(); err != nil {
			return nil, errors.Wrap(err, "new chunk basic header")
		}
		if err = basicHeader.Read(rc.conn); err != nil {
			return nil, errors.Wrap(err, "basic header read")
		}
		log.Tracef("basic header: %s", basicHeader)
		if cs, ok = rc.chunkStreams[basicHeader.chunkStreamID]; !ok {
			if cs, err = newChunkStream(basicHeader.fmt, basicHeader.chunkStreamID); err != nil {
				return nil, err
			}
			rc.chunkStreams[basicHeader.chunkStreamID] = cs
			log.Infof("got a new chunk stream, chunkStreamID: %d", basicHeader.chunkStreamID)
		} else {
			cs.setBasicHeader(basicHeader) //todo 这里容易遗漏！
		}
		if err = cs.readChunk(rc.conn, rc.remoteChunkSize); err != nil {
			return nil, err
		}
		if cs.got() {
			log.Tracef("got chunk stream: %s", cs)
			break
		}
	}

	return cs, nil
}

func (rc *Conn) writeChunkStream(cs *chunkStream) error {
	return cs.writeChunk(rc.conn, rc.chunkSize)
}

func (rc *Conn) handleMessage(chunkStreamID, messageStreamID uint32, messageTypeID uint8, data []byte, timestamp uint32) error {
	log.Tracef("rtmp conn handle message start, chunkStreamID: %d, messageStreamID: %d, messageTypeID: %d, timestamp: %d", chunkStreamID, messageStreamID, messageTypeID, timestamp)
	defer func() {
		log.Tracef("rtmp conn handle message end, chunkStreamID: %d, messageStreamID: %d, messageTypeID: %d, timestamp: %d", chunkStreamID, messageStreamID, messageTypeID, timestamp)
	}()
	switch messageTypeID {
	case typeSetChunkSize,
		typeAbort,
		typeAcknowledgement,
		typeWindowAcknowledgementSize,
		typeSetPeerBandwidth:
		return rc.handleProtocolControlMessage(messageTypeID, data)
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
		return rc.handleCommandMessage(chunkStreamID, messageStreamID, messageTypeID, data, timestamp)
	case typeAggregate: //todo
		return nil
	default:
		return errors.Wrapf(core.ErrorNotSupported, "messageTypeID: %d", messageTypeID)
	}

	return nil
}

func (rc *Conn) handleCommandMessage(chunkStreamID, messageStreamID uint32, typeID uint8, data []byte, timestamp uint32) error {
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
		return rc.handleCommandConnect(chunkStreamID, messageStreamID, vs[1:])
	case commandCall: //todo
	case commandCreateStream:
		return rc.handleCommandCreateStream(chunkStreamID, messageStreamID, vs[1:])
	case commandPlay:
		rc.readMessageDone = true
		return rc.handleCommandPlay(chunkStreamID, messageStreamID, vs[1:])
	case commandPlay2: //todo
	case commandDeleteStream: //todo
	case commandCloseStream: //todo
	case commandReceiveAudio: //todo
	case commandReceiveVideo: //todo
	case commandPublish:
		rc.readMessageDone = true
		rc.isPublish = true
		return rc.handleCommandPublish(chunkStreamID, messageStreamID, vs[1:])
	case commandSeek: //todo
	case commandPause: //todo
	case commandOnStatus: //todo
	default:
		log.Warnf("unknown command: %s", command)
	}
	return nil
}

func (rc *Conn) handleProtocolControlMessage(typeID uint8, data []byte) error {
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
		rc.remoteChunkSize = v
	case typeAbort:
		//todo
		return core.ErrorNotImplemented
	case typeAcknowledgement:
		//todo
		return core.ErrorNotImplemented
	case typeWindowAcknowledgementSize:
		if n != 4 {
			return fmt.Errorf("window acknowledgement size error, length: %d", n)
		}
		rc.remoteWindowAckSize = binary.BigEndian.Uint32(data)
	case typeSetPeerBandwidth:
		//todo
		return core.ErrorNotImplemented
	default:
		return core.ErrorNotSupported
	}
	return nil
}

func (rc *Conn) handleCommandConnect(chunkStreamID, messageStreamID uint32, vs []interface{}) error {
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
			rc.connInfo.App = v.(string)
		}
		if v, ok := obj["flashver"]; ok {
			rc.connInfo.Flashver = v.(string)
		}
		if v, ok := obj["swfUrl"]; ok {
			rc.connInfo.SwfUrl = v.(string)
		}
		if v, ok := obj["tcUrl"]; ok {
			rc.connInfo.TcUrl = v.(string)
		}
		if v, ok := obj["fpad"]; ok {
			rc.connInfo.Fpad = v.(bool)
		}
		if v, ok := obj["audioCodecs"]; ok {
			rc.connInfo.AudioCodecs = int(v.(float64))
		}
		if v, ok := obj["videoCodecs"]; ok {
			rc.connInfo.VideoCodecs = int(v.(float64))
		}
		if v, ok := obj["videoFunction"]; ok {
			rc.connInfo.VideoFunction = int(v.(float64))
		}
		if v, ok := obj["pageUrl"]; ok {
			rc.connInfo.PageUrl = v.(string)
		}
		if v, ok := obj["objectEncoding"]; ok {
			rc.connInfo.ObjectEncoding = v.(float64)
		}
		if v, ok := obj["type"]; ok {
			rc.connInfo.Type = v.(string)
		}
	}
	log.Tracef("handle command connect, connInfo: %+v", rc.connInfo)

	if cs, err := newPCMWindowAcknowledgementSize(2.5e6); err != nil { //todo 为什么是2.5e6
		return err
	} else {
		if err := rc.writeChunkStream(cs); err != nil {
			return err
		}
	}
	if cs, err := newPCMSetPeerBandwidth(2.5e6); err != nil { //todo 为什么是2.5e6
		return err
	} else {
		if err := rc.writeChunkStream(cs); err != nil {
			return err
		}
	}
	if cs, err := newPCMSetChunkSize(rc.chunkSize); err != nil { //todo 是rc.chunkSize吗？
		return err
	} else {
		if err := rc.writeChunkStream(cs); err != nil {
			return err
		}
	}
	if msg, err := newNetConnectionConnectResp(transactionID, rc.connInfo.ObjectEncoding); err != nil {
		return err
	} else {
		if cs, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := rc.writeChunkStream(cs); err != nil {
				return err
			}
		}
	}

	return nil
}

func (rc *Conn) handleCommandCreateStream(chunkStreamID, messageStreamID uint32, vs []interface{}) error {
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
		if cs, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := rc.writeChunkStream(cs); err != nil {
				return err
			}
		}
	}

	return nil
}

func (rc *Conn) handleCommandPublish(chunkStreamID, messageStreamID uint32, vs []interface{}) error {
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
	rc.streamName = publishingName

	if msg, err := newNetStreamPublishResp(); err != nil {
		return err
	} else {
		if cs, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := rc.writeChunkStream(cs); err != nil {
				return err
			}
		}
	}

	return nil
}

func (rc *Conn) handleCommandPlay(chunkStreamID, messageStreamID uint32, vs []interface{}) error {
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
	rc.streamName = streamName

	if cs, err := newUCMStreamIsRecorded(messageStreamID); err != nil {
		return err
	} else {
		if err := rc.writeChunkStream(cs); err != nil {
			return err
		}
	}

	if cs, err := newUCMStreamBegin(messageStreamID); err != nil {
		return err
	} else {
		if err := rc.writeChunkStream(cs); err != nil {
			return err
		}
	}

	if msg, err := newNetStreamPlayReset(); err != nil {
		return err
	} else {
		if cs, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := rc.writeChunkStream(cs); err != nil {
				return err
			}
		}
	}

	if msg, err := newNetStreamPlayStart(); err != nil {
		return err
	} else {
		if cs, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := rc.writeChunkStream(cs); err != nil {
				return err
			}
		}
	}

	//todo ??为什么需要这个
	if msg, err := newNetStreamDataStart(); err != nil {
		return err
	} else {
		if cs, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := rc.writeChunkStream(cs); err != nil {
				return err
			}
		}
	}

	//todo ??为什么需要这个
	if msg, err := newNetStreamPlayPublishNotify(); err != nil {
		return err
	} else {
		if cs, err := newCommandMessage(chunkStreamID, messageStreamID, msg); err != nil {
			return err
		} else {
			if err := rc.writeChunkStream(cs); err != nil {
				return err
			}
		}
	}

	return nil
}
