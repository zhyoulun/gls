package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/zhyoulun/gls/src/server"
	"github.com/zhyoulun/gls/src/stream"
	"github.com/zhyoulun/gls/src/utils/debug"
	"net"
	"os"
)

func Init() error {
	if err := InitLog(); err != nil {
		return err
	}
	if err := debug.Init(); err != nil {
		return err
	}
	if err := stream.InitManager(); err != nil {
		return err
	}
	return nil
}

func InitLog() error {
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
	return nil
}

func main() {
	if err := Init(); err != nil {
		panic(err)
	}

	log.Infof("start golang live server")
	ln, err := net.Listen("tcp", "127.0.0.1:1935")
	if err != nil {
		log.Fatalf("net Listen error: %+v", err)
	}
	rtmpServer, err := server.NewServer(ln)
	if err != nil {
		log.Fatalf("rtmp NewServer error: %+v", err)
	}
	if err := rtmpServer.Serve(); err != nil {
		log.Fatalf("rtmpServer Serve error: %+v", err)
	}
}
