package amf

import "io"

type amf3 struct {
}

func newAmf3() (*amf3, error) {
	return &amf3{}, nil
}

func (a *amf3) decode(r io.Reader) (interface{}, error) {
	//todo
	return nil, nil
}

func (a *amf3) encode(w io.Writer, val interface{}) (int, error) {
	return 0, nil
}
