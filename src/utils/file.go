package utils

import "os"

var ChunkStreamLogFile *os.File
var ChunkLogFile *os.File
var PacketLogFile *os.File

var ChunkStreamHeaderDone bool
var ChunkHeaderDone bool
var PacketHeaderDone bool

func init() {
	var err error
	ChunkStreamLogFile, err = os.OpenFile("temp/chunk_stream_log.csv", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	ChunkLogFile, err = os.OpenFile("temp/chunk_log.csv", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	PacketLogFile, err = os.OpenFile("temp/packet_log.csv", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
}
