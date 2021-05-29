package amf

import "encoding/json"

type AmfObject map[string]interface{}
type AmfArray []interface{}
type AmfTypedObject struct {
	Type   string
	Object AmfObject
}

func (a AmfObject) String() string {
	s, _ := json.Marshal(a)
	return string(s)
}
