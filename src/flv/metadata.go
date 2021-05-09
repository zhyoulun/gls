package flv

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/zhyoulun/gls/src/amf"
)

const (
	setDataFrame = "@setDataFrame"
)

var setDataFrameEncoded []byte

func init() {
	buf := &bytes.Buffer{}
	a, err := amf.NewAmf()
	if err != nil {
		panic(err)
	}
	if _, err := a.Encode(buf, setDataFrame, amf.Amf0); err != nil {
		panic(err)
	}
	setDataFrameEncoded = buf.Bytes()
}

func MetadataReformDelete(data []byte) ([]byte, error) {
	a, err := amf.NewAmf()
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	v, err := a.Decode(r, amf.Amf0)
	if err != nil {
		return nil, err
	}
	switch v.(type) {
	case string:
	default:
		return nil, errors.Errorf("metadata reform, invalid decode")
	}
	vv := v.(string)
	if vv == setDataFrame {
		data = data[len(setDataFrameEncoded):]
	}
	return data, nil
}
