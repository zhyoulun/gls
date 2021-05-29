package rtmp

import (
	"bytes"
	"github.com/zhyoulun/gls/src/amf"
)

func newNetConnectionResponseConnect(transactionID float64) ([]byte, error) {
	properties := make(amf.AmfObject)
	properties["fmsVer"] = "FMS/3,0,1,123"
	properties["capabilities"] = 31

	info := make(amf.AmfObject)
	info["level"] = "status"
	info["code"] = "NetConnection.Connect.Success"
	info["description"] = "Connection Succeeded."
	info["objectEncoding"] = amf.Amf0

	command := amf.AmfArray{
		"_result",
		transactionID,
		properties,
		info,
	}
	return newNetConnectionResponseBase(command)
}

func newNetConnectionResponseCreateStream(transactionID float64, messageStreamID uint32) ([]byte, error) {
	command := amf.AmfArray{
		"_result",
		transactionID,
		nil,
		messageStreamID,
	}
	return newNetConnectionResponseBase(command)
}

func newNetConnectionResponseBase(arr amf.AmfArray) ([]byte, error) {
	a, err := amf.NewAmf()
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if _, err := a.EncodeBatch(buf, arr, amf.Amf0); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
