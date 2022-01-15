package stream

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhyoulun/gls/src/av"
	"github.com/zhyoulun/gls/src/flv"
	"github.com/zhyoulun/gls/src/rtmp"
)

type DataHandler interface {
	ReceiveData(p *av.Packet)
	CloseAllSink()
}

type Source struct {
	handler DataHandler

	conn    *rtmp.Conn
	running bool

	metadata *av.Packet
	video    *av.Packet
	audio    *av.Packet
}

func (s *Source) Run() {
	s.running = true
	go s.readCycle() //todo 待优化到协程池
}

func (s *Source) GetRunning() bool {
	return s.running
}

func (s *Source) GetMetadata() *av.Packet {
	return s.metadata
}

func (s *Source) GetVideo() *av.Packet {
	return s.video
}

func (s *Source) GetAudio() *av.Packet {
	return s.audio
}

func (s *Source) readCycle() {
	log.Debugf("start source")
	for {
		if !s.running {
			break
		}

		var p *av.Packet
		var err error
		if p, err = s.conn.ReadPacket(); err != nil {
			log.Warnf("stream source ReadPacket err: %s", err)
			log.Infof("source conn stop, local addr: %s, remote addr: %s",
				s.conn.NetConn().LocalAddr(), s.conn.NetConn().RemoteAddr())
			s.running = false
			continue
		} else {
			log.Tracef("read %s", p)
		}

		//缓存
		if p.IsMetadata() {
			s.metadata = p
		}
		if p.IsAudio() {
			ah := p.GetAudioTagHandler()
			if ah.SoundFormat() == flv.SoundFormatAAC && ah.AACPacketType() == flv.AACPacketTypeAACSequenceHeader { //todo ??
				s.audio = p
			}
		}
		if p.IsVideo() {
			vh := p.GetVideoTagHandler()
			if vh.FrameType() == flv.FrameTypeKeyFrame && vh.AVCPacketType() == flv.AVCPacketTypeAVCSequenceHeader { //todo ??
				s.video = p
			}
		}

		s.handler.ReceiveData(p)
	}
	log.Debugf("stop source")

	s.handler.CloseAllSink()
}

func NewSource(handler DataHandler, conn *rtmp.Conn) *Source {
	return &Source{
		handler: handler,
		conn:    conn,
		running: false,
	}
}
