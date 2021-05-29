package server

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/zhyoulun/gls/src/rtmp"
	"github.com/zhyoulun/gls/src/stream"
	"github.com/zhyoulun/gls/src/utils"
	"net"
)

type Server struct {
	ln net.Listener
}

func NewServer(ln net.Listener) (*Server, error) {
	server := &Server{
		ln: ln,
	}
	return server, nil
}

func (s *Server) Serve() error {
	for {
		tcpConn, err := s.ln.Accept()
		if err != nil {
			return err
		}
		go func() {
			bufferedConn := utils.NewBufferedConn(tcpConn, 4*1024)
			if err := s.handleConn(bufferedConn); err != nil {
				log.Printf("handle conn error: %+v", err)
			}
		}()
	}
}

func (s *Server) handleConn(conn utils.PeekerConn) error {
	log.Infof("tcp info, local addr: %s, remote addr: %s", conn.LocalAddr(), conn.RemoteAddr())
	rtmpConn, err := rtmp.NewConn(conn)
	if err != nil {
		return err
	}
	if err := rtmpConn.Handshake(); err != nil {
		return err
	}
	if err := rtmpConn.ReadHeader(); err != nil {
		return errors.Wrap(err, "rtmp conn read message")
	}

	if rtmpConn.IsPublish() {
		if err := stream.Mgr.HandlePublish(rtmpConn); err != nil {
			return err
		}
	} else {
		if err := stream.Mgr.HandlePlay(rtmpConn); err != nil {
			return err
		}
	}

	return nil
}
