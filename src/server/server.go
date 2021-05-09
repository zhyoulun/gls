package server

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/zhyoulun/gls/src/rtmp"
	"github.com/zhyoulun/gls/src/stream"
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
			//bufferedConn := utils.NewBufferedConn(tcpConn, 4*1024)
			if err := s.handleConn(tcpConn); err != nil {
				log.Printf("handle conn error: %+v", err)
			}
		}()
	}
}

func (s *Server) handleConn(tcpConn net.Conn) error {
	log.Infof("tcp info, local addr: %s, remote addr: %s", tcpConn.LocalAddr(), tcpConn.RemoteAddr())
	rtmpConn, err := rtmp.NewConn(tcpConn)
	if err != nil {
		return err
	}
	if err := rtmpConn.Handshake(); err != nil {
		return err
	}
	if err := rtmpConn.ReadMessage(); err != nil {
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

	////todo delete
	//if err := tcpConn.Close(); err != nil {
	//	return err
	//}
	//log.Infof("tcp conn close")

	return nil
}
