package core

import (
	"github.com/pkg/errors"
)

var (
	ErrorNotImplemented   = errors.Errorf("not implemented") //协议支持，本项目未实现
	ErrorNotSupported     = errors.Errorf("not supported")   //协议不支持
	ErrorImpossible       = errors.Errorf("impossible error")
	ErrorUnknown          = errors.Errorf("unknown error")
	ErrorAlreadyExist     = errors.Errorf("already exist")
	ErrorDuplicatePublish = errors.Errorf("duplicate publish") //重复推流
	ErrorInvalidData      = errors.Errorf("invalid data")      //不合法的数据
)
