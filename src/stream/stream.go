package stream

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhyoulun/gls/src/av"
	"github.com/zhyoulun/gls/src/core"
	"github.com/zhyoulun/gls/src/rtmp"
	"sync"
)

type Stream struct {
	sinksMutex *sync.Mutex
	sinks      map[string]*Sink

	sourceMutex *sync.Mutex
	source      *Source
}

func newStream() (*Stream, error) {
	return &Stream{
		sinksMutex:  &sync.Mutex{},
		sinks:       make(map[string]*Sink),
		sourceMutex: &sync.Mutex{},
	}, nil
}

func (s *Stream) SetSource(conn *rtmp.Conn) error {
	s.sourceMutex.Lock()
	defer s.sourceMutex.Unlock()

	//重复推流，拒绝后推的流
	if s.source != nil && s.source.GetRunning() {
		if err := conn.Close(); err != nil { //todo close位置优化，减少心智负担
			log.Printf("source Close err: %s", err)
		}
		return core.ErrorDuplicatePublish
	}

	//新建流记录
	s.source = NewSource(s, conn)
	s.source.Run()

	return nil
}

func (s *Stream) ReceiveData(p *av.Packet) {
	s.sendData(p)
}

func (s *Stream) sendData(p *av.Packet) {
	s.sinksMutex.Lock()
	for _, sink := range s.sinks {
		if !sink.initDone {
			if s.source.GetMetadata() != nil {
				if err := sink.Send(s.source.GetMetadata()); err != nil {
					if err := sink.Close(); err != nil {
						log.Errorf("sink Close err: %s", err)
					}
					delete(s.sinks, sink.ID())
					log.Errorf("sink Send err: %s", err)
					continue
				}
			}
			if s.source.GetAudio() != nil {
				if err := sink.Send(s.source.GetAudio()); err != nil {
					if err := sink.Close(); err != nil {
						log.Errorf("sink Close err: %s", err)
					}
					delete(s.sinks, sink.ID())
					log.Errorf("sink Send err: %s", err)
					continue
				}
			}
			if s.source.GetVideo() != nil {
				if err := sink.Send(s.source.GetVideo()); err != nil {
					if err := sink.Close(); err != nil {
						log.Errorf("sink Close err: %s", err)
					}
					delete(s.sinks, sink.ID())
					log.Errorf("sink Send err: %s", err)
					continue
				}
			}
			sink.initDone = true
		}

		if err := sink.Send(p); err != nil {
			if err := sink.Close(); err != nil {
				log.Errorf("sink Close err: %s", err)
			}
			delete(s.sinks, sink.ID())
			log.Errorf("sink Send err: %s", err)
			continue
		}
	}
	s.sinksMutex.Unlock()
}

func (s *Stream) CloseAllSink() {
	s.sinksMutex.Lock()
	defer s.sinksMutex.Unlock()
	for _, sink := range s.sinks {
		err := sink.Close()
		if err != nil {
			log.Warnf("sink Close err: %s", err)
		}
	}
}

func (s *Stream) AddSink(conn *rtmp.Conn) error {
	s.sinksMutex.Lock()
	defer s.sinksMutex.Unlock()
	sink := NewSink(conn)
	s.sinks[sink.ID()] = sink
	sink.Run()
	return nil
}
