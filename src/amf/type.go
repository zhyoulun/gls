package amf

type AmfObject map[string]interface{}
type AmfArray []interface{}
type AmfTypedObject struct {
	Type   string
	Object AmfObject
}
