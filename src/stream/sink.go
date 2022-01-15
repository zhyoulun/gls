package stream

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zhyoulun/gls/src/av"
	"github.com/zhyoulun/gls/src/rtmp"
	"sync/atomic"
)

var (
	globalID int32 = 0
)

type Sink struct {
	id       int32
	conn     *rtmp.Conn
	running  bool
	ch       chan interface{}
	initDone bool
}

func NewSink(conn *rtmp.Conn) *Sink {
	return &Sink{
		id:   atomic.AddInt32(&globalID, 1),
		conn: conn,
		ch:   make(chan interface{}, 1000),
	}
}

func (s *Sink) Close() error {
	s.running = false
	return s.conn.Close()
}

func (s *Sink) Run() {
	s.running = true
	go s.writeCycle() //todo 使用协程池
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
			log.Tracef("write %s", p)
			if err := s.conn.WritePacket(p); err != nil {
				s.running = false
				log.Errorf("write packet fail, err: %s", err)
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
	return fmt.Sprintf("%s:%d", s.conn.GetStreamName(), s.id) //todo 待优化
}
