package amf

import (
	"encoding/binary"
	"fmt"
	"github.com/zhyoulun/gls/src/core"
	"github.com/zhyoulun/gls/src/utils"
	"io"
	"reflect"
)

type amf0 struct {
}

func newAmf0() (*amf0, error) {
	return &amf0{}, nil
}

func (a *amf0) decode(r io.Reader) (interface{}, error) {
	marker, err := a.readMaker(r)
	if err != nil {
		return nil, err
	}
	switch marker {
	case amf0NumberMarker:
		return a.decodeNumber(r)
	case amf0BooleanMarker:
		return a.decodeBoolean(r)
	case amf0StringMarker:
		return a.decodeString(r)
	case amf0ObjectMarker:
		return a.decodeObject(r)
	case amf0MovieclipMarker:
		return nil, core.ErrorNotSupported
	case amf0NullMarker:
		//return nil, a.decodeNull(r)
		return nil, nil
	case amf0UndefinedMarker:
		//return nil, a.decodeUndefined(r)
		return nil, nil
	case amf0ReferenceMarker:
		//todo 协议中有，但livego未实现
		return nil, core.ErrorNotImplemented
	case amf0EcmaArrayMarker:
		return a.decodeEcmaArray(r)
	//object-end-type = UTF-8-empty object-end-marker
	//0x00 0x00 0x09
	//case amf0ObjectEndMarker:
	case amf0StrictArrayMarker:
		return a.decodeStrictArray(r)
	case amf0DateMarker:
		return a.decodeDate(r)
	case amf0LongStringMarker:
		return a.decodeLongString(r)
	case amf0UnsupportedMarker:
		//if a type cannot be serialized a special unsupported marker can be used in place of the type.
		//Some endpoints may throw an error on encountering this type marker.
		//No further information is encoded for this type.
		return nil, core.ErrorNotSupported
	case amf0RecordsetMarker:
		return nil, core.ErrorNotSupported
	case amf0XmlDocumentMarker:
		return a.decodeXmlDocument(r)
	case amf0TypedObjectMarker:
		return a.decodeTypedObject(r)
	case amf0AvmplusObjectMarker:
		if amf3, err := newAmf3(); err != nil {
			return nil, err
		} else {
			return amf3.decode(r)
		}
	}
	return nil, fmt.Errorf("amf0 decode, unknown marker %d", marker)
}

func (a *amf0) encode(w io.Writer, val interface{}) (int, error) {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Invalid:
		return a.encodeNull(w)
	case reflect.Bool:
		return a.encodeBoolean(w, v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.encodeNumber(w, float64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return a.encodeNumber(w, float64(v.Uint()))
	case reflect.Uintptr: //todo?
	case reflect.Float32, reflect.Float64:
		return a.encodeNumber(w, v.Float())
	case reflect.Complex64: //todo?
	case reflect.Complex128: //todo?
	case reflect.Array, reflect.Slice:
		length := v.Len()
		arr := make(AmfArray, 0, length)
		for i := 0; i < length; i++ {
			arr = append(arr, v.Index(i).Interface())
		}
		return a.encodeStrictArray(w, arr)
	case reflect.Chan: //todo?
	case reflect.Func: //todo?
	case reflect.Interface: //todo?
	case reflect.Map:
		if obj, ok := val.(AmfObject); !ok {
			return 0, fmt.Errorf("amf0 encode: unable to create object from map")
		} else {
			return a.encodeObject(w, obj)
		}
	case reflect.Ptr: //todo?
	case reflect.String:
		str := v.String()
		if len(str) <= amf0StringMax {
			return a.encodeString(w, str, true)
		} else {
			return a.encodeLongString(w, str)
		}
	case reflect.Struct:
		//todo AmfTypedObject
	case reflect.UnsafePointer: //todo?
	}

	return 0, fmt.Errorf("amf0 encode: unsupport v kind %v", v.Kind())
}

func (a *amf0) readMaker(r io.Reader) (byte, error) {
	return utils.ReadByte(r)
}

//读取一个期望的marker
func (a *amf0) readWantMarker(r io.Reader, want byte) error {
	got, err := a.readMaker(r)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("amf0 deocde: assert marker, want=%d, got=%d", want, got)
	}
	return nil
}

func (a *amf0) writeMarker(w io.Writer, m byte) error {
	return utils.WriteByte(w, m)
}

//number-type = number-marker DOUBLE
func (a *amf0) decodeNumber(r io.Reader) (float64, error) {
	var n float64
	err := binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

//boolean-type = boolean-marker U8
//0 is false, <>0 is true
func (a *amf0) decodeBoolean(r io.Reader) (bool, error) {
	b, err := utils.ReadByte(r)
	if err != nil {
		return false, err
	}
	if b == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

//string-type = string-marker UTF-8
func (a *amf0) decodeString(r io.Reader) (string, error) {
	var length uint16
	err := binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return "", err
	}

	buf, err := utils.ReadBytes(r, int(length))
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

//object-property = (UTF-8 value-type) | (UTF-8-empty object-end-marker)
//anonymous-object-type = object-marker *(object-property)
func (a *amf0) decodeObject(r io.Reader) (AmfObject, error) {
	result := make(AmfObject) //todo 优化
	for {
		key, err := a.decodeString(r)
		if err != nil {
			return nil, err
		}
		if key == "" {
			err := a.readWantMarker(r, amf0ObjectEndMarker)
			if err != nil {
				return nil, err
			}
			break
		}
		value, err := a.decode(r)
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

//null-type = null-marker
//func (a *amf0) decodeNull(r io.Reader) error {
//	//return a.readWantMarker(r, amf0NullMarker)
//	return nil
//}

//undefined-type = undefined-marker
//func (a *amf0) decodeUndefined(r io.Reader) error {
//	//return a.readWantMarker(r, amf0UndefinedMarker)
//	return nil
//}

//associative-count = U32
//ecma-arrya-type = associative-count *(object-property)
func (a *amf0) decodeEcmaArray(r io.Reader) (AmfObject, error) {
	var length uint32 //useless
	err := binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	return a.decodeObject(r)
}

//array-count = U32
//strict-array-type = array-count *(value-type)
func (a *amf0) decodeStrictArray(r io.Reader) (AmfArray, error) {
	var length uint32
	err := binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0, length)
	for i := int64(0); i < int64(length); i++ {
		item, err := a.decode(r)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

//time-zone = S16;reserved, not supported, should be set to 0x0000
//date-type = date-marker DOUBLE time-zone
//todo 这里不知道为什么，实现和规范不一致。。是反着的
func (a *amf0) decodeDate(r io.Reader) (float64, error) {
	result, err := a.decodeNumber(r)
	if err != nil {
		return float64(0), err
	}

	var n int16 //useless
	err = binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return float64(0), err
	}
	return result, nil
}

//long-string-type = long-string-marker UTF-8-long
func (a *amf0) decodeLongString(r io.Reader) (string, error) {
	var length int32
	err := binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return "", err
	}

	buf, err := utils.ReadBytes(r, int(length))
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

//xml-document-type = xml-document-marker UTF-8-long
func (a *amf0) decodeXmlDocument(r io.Reader) (string, error) {
	return a.decodeLongString(r)
}

//class-name = UTF-8
//object-type = object-marker class-name *(object-property)
func (a *amf0) decodeTypedObject(r io.Reader) (*AmfTypedObject, error) {
	result := new(AmfTypedObject) //todo 优化
	var err error
	result.Type, err = a.decodeString(r)
	if err != nil {
		return nil, err
	}
	result.Object, err = a.decodeObject(r)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *amf0) encodeNull(w io.Writer) (int, error) {
	n := 0
	if err := a.writeMarker(w, amf0NullMarker); err != nil {
		return 0, err
	}
	n += 1
	return n, nil
}

func (a *amf0) encodeBoolean(w io.Writer, b bool) (int, error) {
	n := 0
	if err := a.writeMarker(w, amf0BooleanMarker); err != nil {
		return 0, err
	}
	n += 1
	if b {
		if err := utils.WriteByte(w, amf0BooleanTrue); err != nil {
			return 0, err
		}
	} else {
		if err := utils.WriteByte(w, amf0BooleanFalse); err != nil {
			return 0, err
		}
	}
	n += 1
	return n, nil
}

func (a *amf0) encodeNumber(w io.Writer, num float64) (int, error) {
	n := 0
	if err := a.writeMarker(w, amf0NumberMarker); err != nil {
		return 0, err
	}
	n += 1
	if err := binary.Write(w, binary.BigEndian, &num); err != nil {
		return 0, err
	}
	n += 8
	return n, nil
}

func (a *amf0) encodeStrictArray(w io.Writer, arr AmfArray) (int, error) {
	n := 0
	if err := a.writeMarker(w, amf0StrictArrayMarker); err != nil {
		return 0, err
	}
	n += 1
	length := uint32(len(arr))
	if err := binary.Write(w, binary.BigEndian, &length); err != nil {
		return 0, err
	}
	n += 4
	for _, v := range arr {
		if nn, err := a.encode(w, v); err != nil {
			return 0, err
		} else {
			n += nn
		}
	}
	return n, nil
}

func (a *amf0) encodeObject(w io.Writer, obj AmfObject) (int, error) {
	n := 0
	if err := a.writeMarker(w, amf0ObjectMarker); err != nil {
		return 0, err
	} else {
		n += 1
	}
	for k, v := range obj {
		if nn, err := a.encodeString(w, k, false); err != nil {
			return 0, err
		} else {
			n += nn
		}
		if nn, err := a.encode(w, v); err != nil {
			return 0, err
		} else {
			n += nn
		}
	}
	if nn, err := a.encodeString(w, "", false); err != nil {
		return 0, err
	} else {
		n += nn
	}
	if err := a.writeMarker(w, amf0ObjectEndMarker); err != nil {
		return 0, err
	} else {
		n += 1
	}
	return n, nil
}

func (a *amf0) encodeString(w io.Writer, s string, needMarker bool) (int, error) {
	n := 0
	if needMarker {
		if err := a.writeMarker(w, amf0StringMarker); err != nil {
			return 0, err
		} else {
			n += 1
		}
	}
	length := uint16(len(s))
	if err := binary.Write(w, binary.BigEndian, &length); err != nil {
		return 0, err
	} else {
		n += 2
	}
	if err := utils.WriteBytes(w, []byte(s)); err != nil {
		return 0, err
	} else {
		n += int(length)
	}
	return n, nil
}

func (a *amf0) encodeLongString(w io.Writer, s string) (int, error) {
	n := 0
	if err := a.writeMarker(w, amf0LongStringMarker); err != nil {
		return 0, err
	} else {
		n += 1
	}
	length := int32(len(s))
	if err := binary.Write(w, binary.BigEndian, &length); err != nil {
		return 0, err
	} else {
		n += 4
	}
	if err := utils.WriteBytes(w, []byte(s)); err != nil {
		return 0, err
	} else {
		n += int(length)
	}
	return n, nil
}
