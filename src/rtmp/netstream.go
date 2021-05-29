package rtmp

import (
	"bytes"
	"github.com/zhyoulun/gls/src/amf"
)

type netStreamStatusInfoObject struct {
	level       string //the level for this message
	code        string //the message code
	description string //a human-readable description of the message
	//the info object may contain other properties as appropriate to the code
}

func newNetStreamStatusInfoObject(level, code, description string) *netStreamStatusInfoObject {
	return &netStreamStatusInfoObject{
		level:       level,
		code:        code,
		description: description,
	}
}

func (info *netStreamStatusInfoObject) toAmfObject() amf.AmfObject {
	res := make(amf.AmfObject)
	res["level"] = info.level
	res["code"] = info.code
	res["description"] = info.description
	return res
}

func newNetStreamResponsePublishStart() ([]byte, error) {
	return newNetStreamResponseBase("NetStream.Publish.Start", "Start publishing.")
}

func newNetStreamResponsePlayReset() ([]byte, error) {
	return newNetStreamResponseBase("NetStream.Play.Reset", "Playing and resetting stream.")
}

func newNetStreamResponsePlayStart() ([]byte, error) {
	return newNetStreamResponseBase("NetStream.Play.Start", "Started playing stream.")
}

func newNetStreamResponseBase(code, description string) ([]byte, error) {
	infoObject := newNetStreamStatusInfoObject(netStreamStatusLevelStatus, code, description).toAmfObject()
	command := amf.AmfArray{
		commandNetStreamOnStatus,
		transactionID0,
		nil,
		infoObject,
	}

	a, err := amf.NewAmf()
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if _, err := a.EncodeBatch(buf, command, amf.Amf0); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
