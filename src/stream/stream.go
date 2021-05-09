package stream

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhyoulun/gls/src/av"
	"github.com/zhyoulun/gls/src/core"
	"github.com/zhyoulun/gls/src/flv"
	"github.com/zhyoulun/gls/src/rtmp"
	"sync"
)

type Stream struct {
	sinksMutex  *sync.Mutex
	sinks       []*Sink
	sourceMutex *sync.Mutex
	source      *rtmp.Conn
	running     bool

	metadata *av.Packet
	video    *av.Packet
	audio    *av.Packet
}

func newStream() (*Stream, error) {
	return &Stream{
		sinksMutex:  &sync.Mutex{},
		sinks:       make([]*Sink, 0),
		sourceMutex: &sync.Mutex{},
		source:      nil,
		running:     false,
	}, nil
}

func (s *Stream) SetSource(conn *rtmp.Conn) error {
	s.sourceMutex.Lock()
	defer s.sourceMutex.Unlock()

	if s.source != nil {
		if err := conn.Close(); err != nil { //todo close位置优化，减少心智负担
			log.Printf("source Close error: %+v", err)
		}
		return core.ErrorAlreadyExist
	}
	s.source = conn
	return nil
}

func (s *Stream) Run() {
	s.running = true
	go s.readCycle()
}

func (s *Stream) readCycle() {
	//i := 0
	for {
		if !s.running {
			s.closeAllSink()
			break
		}
		var p *av.Packet
		var err error
		if p, err = s.source.ReadPacket(); err != nil {
			s.running = false
			continue
		} else {
			log.Infof("%s", p)
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

		////todo delete
		//i++
		//if i > 100 {
		//	break
		//}
		//todo write to sink
		s.sinksMutex.Lock()
		for _, sink := range s.sinks {
			if !sink.initDone {
				if s.metadata != nil {
					sink.Write(s.metadata)
				}
				if s.audio != nil {
					sink.Write(s.audio)
				}
				if s.video != nil {
					sink.Write(s.video)
				}
				sink.initDone = true
			}
			sink.Write(p)
		}
		s.sinksMutex.Unlock()
	}
}

func (s *Stream) closeAllSink() {
	//todo
}

func (s *Stream) addSink(conn *rtmp.Conn) error {
	s.sinksMutex.Lock()
	defer s.sinksMutex.Unlock()
	sink, err := newSink(conn)
	if err != nil {
		return err
	}
	s.sinks = append(s.sinks, sink)
	sink.Run()
	return nil
}
