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
		ch:   make(chan interface{}, 1000),
	}, nil
}

func (s *Sink) Close() error {
	s.running = false
	return s.conn.Close()
}

func (s *Sink) Run() {
	s.running = true
	go s.writeCycle()
}

func (s *Sink) writeCycle() {
	log.Infof("sink start cycle")
	for {
		if !s.running {
			break
		}

		select {
		case ch := <-s.ch:
			p := ch.(*av.Packet)
			log.Infof("write %s", p)
			if err := s.conn.WritePacket(p); err != nil {
				s.running = false
				log.Errorf("write packet fail, err: %+v", err)
			}
		}
	}
	log.Infof("sink end cycle")
}

func (s *Sink) Send(p *av.Packet) error {
	select {
	case s.ch <- p:
	default:
		select {
		case s.ch <- p:
			//default:
			//	return errors.Errorf("sink[%s] send to channel fail", s.ID())
		}
	}
	return nil
}

func (s *Sink) ID() string {
	return s.conn.GetStreamName() //todo 待优化
}
