package stream

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhyoulun/gls/src/av"
	"github.com/zhyoulun/gls/src/rtmp"
)

type Sink struct {
	conn     *rtmp.Conn
	running  bool
	ch       chan interface{}
	initDone bool
}

func newSink(conn *rtmp.Conn) (*Sink, error) {
	return &Sink{
		conn: conn,
		ch:   make(chan interface{}),
	}, nil
}

func (s *Sink) Run() {
	s.running = true
	go s.writeCycle()
}

func (s *Sink) writeCycle() {
	for {
		if !s.running {
			break
		}

		select {
		case ch := <-s.ch:
			p := ch.(*av.Packet)
			log.Infof("get one packet to write, %s", p)
			if err := s.conn.WritePacket(p); err != nil {
				s.running = false
				log.Errorf("write packet fail, err: %+v", err)
			}
		default:
		}
	}
}

func (s *Sink) Write(p *av.Packet) {
	s.ch <- p
}
