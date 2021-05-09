package rtmp

import (
	"bytes"
	"github.com/zhyoulun/gls/src/amf"
)

func newNetConnectionConnectResp(transactionID float64, encoding float64) ([]byte, error) {
	properties := make(amf.AmfObject)
	properties["fmsVer"] = "FMS/3,0,1,123"
	properties["capabilities"] = 31

	information := make(amf.AmfObject)
	information["level"] = "status"
	information["code"] = "NetConnection.Connect.Success"
	information["description"] = "Connection Succeeded."
	information["objectEncoding"] = encoding

	command := amf.AmfArray{
		"_result",
		transactionID,
		properties,
		information,
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

func newCreateStreamResp(transactionID float64, messageStreamID uint32) ([]byte, error) {
	command := amf.AmfArray{
		"_result",
		transactionID,
		nil,
		messageStreamID,
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

func newNetStreamPublishResp() ([]byte, error) {
	info := make(amf.AmfObject)
	info["level"] = "status"
	info["code"] = "NetStream.Publish.Start"
	info["description"] = "Start publishing."
	command := amf.AmfArray{
		"onStatus",
		0,
		nil,
		info,
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

func newNetStreamPlayReset() ([]byte, error) {
	info := make(amf.AmfObject)
	info["level"] = "status"
	info["code"] = "NetStream.Play.Reset"
	info["description"] = "Playing and resetting stream."
	command := amf.AmfArray{
		"onStatus",
		0,
		nil,
		info,
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

func newNetStreamPlayStart() ([]byte, error) {
	info := make(amf.AmfObject)
	info["level"] = "status"
	info["code"] = "NetStream.Play.Start"
	info["description"] = "Started playing stream."
	command := amf.AmfArray{
		"onStatus",
		0,
		nil,
		info,
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

//todo ??为什么需要这个
func newNetStreamDataStart() ([]byte, error) {
	info := make(amf.AmfObject)
	info["level"] = "status"
	info["code"] = "NetStream.Data.Start"
	info["description"] = "Started playing stream."
	command := amf.AmfArray{
		"onStatus",
		0,
		nil,
		info,
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

//todo ??为什么需要这个
func newNetStreamPlayPublishNotify() ([]byte, error) {
	info := make(amf.AmfObject)
	info["level"] = "status"
	info["code"] = "NetStream.Play.PublishNotify"
	info["description"] = "Started playing notify."
	command := amf.AmfArray{
		"onStatus",
		0,
		nil,
		info,
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
