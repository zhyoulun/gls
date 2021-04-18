package core

import "fmt"

var (
	ErrorNotImplemented = fmt.Errorf("not implemented")
	ErrorNotSupported   = fmt.Errorf("not supported")
)
