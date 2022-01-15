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

	addr := "127.0.0.1:1935"
	log.Infof("start golang live server: %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("net Listen err: %s", err)
	}
	rtmpServer, err := server.NewServer(ln)
	if err != nil {
		log.Fatalf("rtmp NewServer err: %s", err)
	}
	if err := rtmpServer.Serve(); err != nil {
		log.Fatalf("rtmpServer Serve err: %s", err)
	}
}
