package amf

import (
	"github.com/zhyoulun/gls/src/core"
	"io"
)

type Amf struct {
	amf0 *amf0
	amf3 *amf3
}

func NewAmf() (*Amf, error) {
	amf0, err := newAmf0()
	if err != nil {
		return nil, err
	}
	amf3, err := newAmf3()
	if err != nil {
		return nil, err
	}
	return &Amf{amf0: amf0, amf3: amf3}, nil
}

func (a *Amf) Encode(w io.Writer, val interface{}, ver AmfVersion) (int, error) {
	switch ver {
	case Amf0:
		return a.amf0.encode(w, val)
	case Amf3:
		return a.amf3.encode(w, val)
	}
	return 0, core.ErrorNotSupported
}

func (a *Amf) Decode(r io.Reader, ver AmfVersion) (interface{}, error) {
	switch ver {
	case Amf0:
		return a.amf0.decode(r)
	case Amf3:
		return a.amf3.decode(r)
	}
	return 0, core.ErrorNotSupported
}
