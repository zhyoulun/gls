package amf

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"github.com/zhyoulun/gls/src/utils"
	"strings"
	"testing"
)

func Test_amf0_decodeNumber(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, 1.23)
		num, err := a.decodeNumber(buf)
		assert.Equal(t, num, 1.23)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		num, err := a.decodeNumber(buf)
		assert.Equal(t, num, 0.0)
		assert.Error(t, err)
	}
}

func Test_amf0_decodeBoolean(t *testing.T) {
	a, _ := newAmf0()
	{
		r := strings.NewReader("abc")
		b, err := a.decodeBoolean(r)
		assert.True(t, b)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, int16(1))
		b, err := a.decodeBoolean(buf)
		assert.False(t, b)
		assert.NoError(t, err)
	}
	{
		r := strings.NewReader("")
		b, err := a.decodeBoolean(r)
		assert.False(t, b)
		assert.Error(t, err)
	}
}

func Test_amf0_decodeString(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, uint16(4))
		buf.Write([]byte(`abcd`))
		s, err := a.decodeString(buf)
		assert.Equal(t, "abcd", s)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		s, err := a.decodeString(buf)
		assert.Equal(t, s, "")
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, uint16(4))
		s, err := a.decodeString(buf)
		assert.Equal(t, s, "")
		assert.Error(t, err)
	}
}

func Test_amf0_decodeObject(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		//write key
		binary.Write(buf, binary.BigEndian, uint16(4))
		buf.Write([]byte(`abcd`))
		//write value
		buf.WriteByte(0x01)
		buf.WriteByte(0x01)
		//write empty key string
		binary.Write(buf, binary.BigEndian, uint16(0))
		//write object end marker
		buf.WriteByte(0x09)
		m, err := a.decodeObject(buf)
		assert.Equal(t, m, AmfObject{
			"abcd": true,
		})
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		//write key
		binary.Write(buf, binary.BigEndian, uint16(4))
		buf.Write([]byte(`abcd`))
		//write value
		buf.WriteByte(0x01)
		buf.WriteByte(0x01)
		//write empty key string
		binary.Write(buf, binary.BigEndian, uint16(0))
		//write object end marker
		//buf.WriteByte(0x09)
		m, err := a.decodeObject(buf)
		assert.Nil(t, m)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		//write key
		binary.Write(buf, binary.BigEndian, uint16(4))
		buf.Write([]byte(`abcd`))
		//write value
		buf.WriteByte(0x01)
		buf.WriteByte(0x01)
		//write empty key string
		binary.Write(buf, binary.BigEndian, uint16(0))
		//write object end marker
		buf.WriteByte(0x19)
		m, err := a.decodeObject(buf)
		assert.Nil(t, m)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		//write key
		binary.Write(buf, binary.BigEndian, uint16(4))
		buf.Write([]byte(`abcd`))
		//write value
		buf.WriteByte(0x01)
		//buf.WriteByte(0x01)
		////write empty key string
		//binary.Write(buf, binary.BigEndian, uint16(0))
		////write object end marker
		//buf.WriteByte(0x09)
		m, err := a.decodeObject(buf)
		assert.Nil(t, m)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		m, err := a.decodeObject(buf)
		assert.Nil(t, m)
		assert.Error(t, err)
	}
}

func Test_amf0_readWantMarker(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		err := a.readWantMarker(buf, amf0NullMarker)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0ObjectEndMarker)
		err := a.readWantMarker(buf, amf0NullMarker)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0ObjectEndMarker)
		err := a.readWantMarker(buf, amf0ObjectEndMarker)
		assert.NoError(t, err)
	}
}

//func Test_amf0_decodeNull(t *testing.T) {
//	a, _ := newAmf0()
//	{
//		buf := &bytes.Buffer{}
//		buf.WriteByte(amf0NullMarker)
//		err := a.decodeNull(buf)
//		assert.NoError(t, err)
//	}
//	{
//		buf := &bytes.Buffer{}
//		buf.WriteByte(amf0ObjectEndMarker)
//		err := a.decodeNull(buf)
//		assert.Error(t, err)
//	}
//}

//func Test_amf0_decodeUndefined(t *testing.T) {
//	a, _ := newAmf0()
//	{
//		buf := &bytes.Buffer{}
//		buf.WriteByte(amf0UndefinedMarker)
//		err := a.decodeUndefined(buf)
//		assert.NoError(t, err)
//	}
//	{
//		buf := &bytes.Buffer{}
//		buf.WriteByte(amf0ObjectEndMarker)
//		err := a.decodeUndefined(buf)
//		assert.Error(t, err)
//	}
//}

func Test_amf0_decodeEcmaArray(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		//write useless uint32
		binary.Write(buf, binary.BigEndian, uint32(1))
		//write key
		binary.Write(buf, binary.BigEndian, uint16(4))
		buf.Write([]byte(`abcd`))
		//write value
		buf.WriteByte(0x01)
		buf.WriteByte(0x01)
		//write empty key string
		binary.Write(buf, binary.BigEndian, uint16(0))
		//write object end marker
		buf.WriteByte(0x09)
		m, err := a.decodeEcmaArray(buf)
		assert.Equal(t, m, AmfObject{
			"abcd": true,
		})
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		m, err := a.decodeEcmaArray(buf)
		assert.Nil(t, m)
		assert.Error(t, err)
	}
}

func Test_amf0_decodeStrictArray(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		//write uint32
		binary.Write(buf, binary.BigEndian, uint32(1))
		//write value
		buf.WriteByte(0x01)
		buf.WriteByte(0x01)
		//write empty key string
		binary.Write(buf, binary.BigEndian, uint16(0))
		//write object end marker
		buf.WriteByte(0x09)
		m, err := a.decodeStrictArray(buf)
		assert.Equal(t, AmfArray{true}, m)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		//write uint32
		binary.Write(buf, binary.BigEndian, uint32(2))
		//write value
		buf.WriteByte(0x01)
		buf.WriteByte(0x01)
		////write empty key string
		//binary.Write(buf, binary.BigEndian, uint16(0))
		////write object end marker
		//buf.WriteByte(0x09)
		m, err := a.decodeStrictArray(buf)
		assert.Nil(t, m)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		m, err := a.decodeStrictArray(buf)
		assert.Nil(t, m)
		assert.Error(t, err)
	}
}

func Test_amf0_decodeDate(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, 1.23)
		binary.Write(buf, binary.BigEndian, uint16(1))
		num, err := a.decodeDate(buf)
		assert.Equal(t, 1.23, num)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, 1.23)
		num, err := a.decodeDate(buf)
		assert.Equal(t, 0.0, num)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		num, err := a.decodeDate(buf)
		assert.Equal(t, 0.0, num)
		assert.Error(t, err)
	}
}

func Test_amf0_decodeLongString(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, int32(4))
		buf.Write([]byte(`abcd`))
		s, err := a.decodeLongString(buf)
		assert.Equal(t, "abcd", s)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, int32(4))
		s, err := a.decodeLongString(buf)
		assert.Equal(t, "", s)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		s, err := a.decodeLongString(buf)
		assert.Equal(t, "", s)
		assert.Error(t, err)
	}
}

func Test_amf0_decodeXmlDocument(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, int32(4))
		buf.Write([]byte(`abcd`))
		s, err := a.decodeXmlDocument(buf)
		assert.Equal(t, "abcd", s)
		assert.NoError(t, err)
	}
}

func Test_amf0_decodeTypedObject(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, int16(4))
		buf.Write([]byte(`abcd`))
		//write key
		binary.Write(buf, binary.BigEndian, uint16(4))
		buf.Write([]byte(`abcd`))
		//write value
		buf.WriteByte(0x01)
		buf.WriteByte(0x01)
		//write empty key string
		binary.Write(buf, binary.BigEndian, uint16(0))
		//write object end marker
		buf.WriteByte(0x09)
		s, err := a.decodeTypedObject(buf)
		assert.Equal(t, &AmfTypedObject{"abcd", AmfObject{"abcd": true}}, s)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, int16(4))
		buf.Write([]byte(`abcd`))
		////write key
		//binary.Write(buf, binary.BigEndian, uint16(4))
		//buf.Write([]byte(`abcd`))
		////write value
		//buf.WriteByte(0x01)
		//buf.WriteByte(0x01)
		////write empty key string
		//binary.Write(buf, binary.BigEndian, uint16(0))
		////write object end marker
		//buf.WriteByte(0x09)
		s, err := a.decodeTypedObject(buf)
		assert.Nil(t, s)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		//binary.Write(buf, binary.BigEndian, int16(4))
		//buf.Write([]byte(`abcd`))
		////write key
		//binary.Write(buf, binary.BigEndian, uint16(4))
		//buf.Write([]byte(`abcd`))
		////write value
		//buf.WriteByte(0x01)
		//buf.WriteByte(0x01)
		////write empty key string
		//binary.Write(buf, binary.BigEndian, uint16(0))
		////write object end marker
		//buf.WriteByte(0x09)
		s, err := a.decodeTypedObject(buf)
		assert.Nil(t, s)
		assert.Error(t, err)
	}
}

func Test_amf0_decode(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0NumberMarker)
		binary.Write(buf, binary.BigEndian, 1.23)
		num, err := a.decode(buf)
		assert.Equal(t, num, 1.23)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0BooleanMarker)
		buf.Write([]byte(`abc`))
		b, err := a.decode(buf)
		assert.Equal(t, true, b)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0StringMarker)
		binary.Write(buf, binary.BigEndian, uint16(4))
		buf.Write([]byte(`abcd`))
		s, err := a.decode(buf)
		assert.Equal(t, s, "abcd")
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0ObjectMarker)
		//write key
		binary.Write(buf, binary.BigEndian, uint16(4))
		buf.Write([]byte(`abcd`))
		//write value
		buf.WriteByte(0x01)
		buf.WriteByte(0x01)
		//write empty key string
		binary.Write(buf, binary.BigEndian, uint16(0))
		//write object end marker
		buf.WriteByte(0x09)
		m, err := a.decode(buf)
		assert.Equal(t, m, AmfObject{
			"abcd": true,
		})
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0NullMarker)
		s, err := a.decode(buf)
		assert.Equal(t, nil, s)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0UndefinedMarker)
		s, err := a.decode(buf)
		assert.Equal(t, nil, s)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0MovieclipMarker)
		s, err := a.decode(buf)
		assert.Equal(t, nil, s)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0ReferenceMarker)
		s, err := a.decode(buf)
		assert.Equal(t, nil, s)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0EcmaArrayMarker)
		//write useless uint32
		binary.Write(buf, binary.BigEndian, uint32(1))
		//write key
		binary.Write(buf, binary.BigEndian, uint16(4))
		buf.Write([]byte(`abcd`))
		//write value
		buf.WriteByte(0x01)
		buf.WriteByte(0x01)
		//write empty key string
		binary.Write(buf, binary.BigEndian, uint16(0))
		//write object end marker
		buf.WriteByte(0x09)
		m, err := a.decode(buf)
		assert.Equal(t, m, AmfObject{
			"abcd": true,
		})
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0StrictArrayMarker)
		//write uint32
		binary.Write(buf, binary.BigEndian, uint32(1))
		//write value
		buf.WriteByte(0x01)
		buf.WriteByte(0x01)
		//write empty key string
		binary.Write(buf, binary.BigEndian, uint16(0))
		//write object end marker
		buf.WriteByte(0x09)
		m, err := a.decode(buf)
		assert.Equal(t, AmfArray{true}, m)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0DateMarker)
		binary.Write(buf, binary.BigEndian, 1.23)
		binary.Write(buf, binary.BigEndian, uint16(1))
		num, err := a.decode(buf)
		assert.Equal(t, 1.23, num)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0LongStringMarker)
		binary.Write(buf, binary.BigEndian, int32(4))
		buf.Write([]byte(`abcd`))
		s, err := a.decode(buf)
		assert.Equal(t, "abcd", s)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0UnsupportedMarker)
		s, err := a.decode(buf)
		assert.Equal(t, nil, s)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0RecordsetMarker)
		s, err := a.decode(buf)
		assert.Equal(t, nil, s)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0XmlDocumentMarker)
		binary.Write(buf, binary.BigEndian, int32(4))
		buf.Write([]byte(`abcd`))
		s, err := a.decode(buf)
		assert.Equal(t, "abcd", s)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0TypedObjectMarker)
		binary.Write(buf, binary.BigEndian, int16(4))
		buf.Write([]byte(`abcd`))
		//write key
		binary.Write(buf, binary.BigEndian, uint16(4))
		buf.Write([]byte(`abcd`))
		//write value
		buf.WriteByte(0x01)
		buf.WriteByte(0x01)
		//write empty key string
		binary.Write(buf, binary.BigEndian, uint16(0))
		//write object end marker
		buf.WriteByte(0x09)
		s, err := a.decode(buf)
		assert.Equal(t, &AmfTypedObject{"abcd", AmfObject{"abcd": true}}, s)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(amf0AvmplusObjectMarker)
		s, err := a.decode(buf)
		assert.Equal(t, nil, s)
		assert.NoError(t, err)
	}
	{
		buf := &bytes.Buffer{}
		buf.WriteByte(0xff)
		s, err := a.decode(buf)
		assert.Equal(t, nil, s)
		assert.Error(t, err)
	}
	{
		buf := &bytes.Buffer{}
		s, err := a.decode(buf)
		assert.Equal(t, nil, s)
		assert.Error(t, err)
	}
}

func TestTmp(t *testing.T) {
	//a, _ := newAmf0()
	//{
	//	buf := &bytes.Buffer{}
	//	buf.WriteByte(amf0LongStringMarker)
	//	binary.Write(buf, binary.BigEndian, int32(4))
	//	buf.Write([]byte(`abcd`))
	//	s, err := a.decodeLongString(buf)
	//	assert.Equal(t, "abcd", s)
	//	assert.NoError(t, err)
	//}
}

func Test_amf0_writeMarker(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		err := a.writeMarker(buf, amf0StringMarker)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0StringMarker}, buf.Bytes())
	}
}

func Test_amf0_encodeNull(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		n, err := a.encodeNull(buf)
		assert.Equal(t, 1, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0NullMarker}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(0)
		n, err := a.encodeNull(buf)
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{}, buf.Bytes())
	}
}

func Test_amf0_encodeBoolean(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		n, err := a.encodeBoolean(buf, true)
		assert.Equal(t, 2, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0BooleanMarker, amf0BooleanTrue}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		n, err := a.encodeBoolean(buf, false)
		assert.Equal(t, 2, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0BooleanMarker, amf0BooleanFalse}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(0)
		n, err := a.encodeBoolean(buf, true)
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(1)
		n, err := a.encodeBoolean(buf, true)
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{amf0BooleanMarker}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(1)
		n, err := a.encodeBoolean(buf, false)
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{amf0BooleanMarker}, buf.Bytes())
	}
}

func Test_amf0_encodeNumber(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		n, err := a.encodeNumber(buf, 0.0)
		assert.Equal(t, 9, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0NumberMarker, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(1)
		n, err := a.encodeNumber(buf, 1.23)
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{amf0NumberMarker}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(0)
		n, err := a.encodeNumber(buf, 1.23)
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{}, buf.Bytes())
	}
}

func Test_amf0_encodeStrictArray(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		n, err := a.encodeStrictArray(buf, AmfArray{})
		assert.Equal(t, 5, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0StrictArrayMarker, 0x00, 0x00, 0x00, 0x00}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		n, err := a.encodeStrictArray(buf, AmfArray{true})
		assert.Equal(t, 7, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0StrictArrayMarker, 0x00, 0x00, 0x00, 0x01, amf0BooleanMarker, amf0BooleanTrue}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(0)
		n, err := a.encodeStrictArray(buf, AmfArray{})
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(3)
		n, err := a.encodeStrictArray(buf, AmfArray{})
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{amf0StrictArrayMarker}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(5)
		n, err := a.encodeStrictArray(buf, AmfArray{true})
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{amf0StrictArrayMarker, 0x00, 0x00, 0x00, 0x01}, buf.Bytes())
	}
}

func Test_amf0_encodeObject(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		n, err := a.encodeObject(buf, AmfObject{})
		assert.Equal(t, 4, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0ObjectMarker, 0x00, 0x00, amf0ObjectEndMarker}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		n, err := a.encodeObject(buf, AmfObject{"a": true})
		assert.Equal(t, 9, n)
		assert.NoError(t, err)
		want := []byte{amf0ObjectMarker,
			0x00, 0x01, 'a',
			amf0BooleanMarker, amf0BooleanTrue,
			0x00, 0x00, amf0ObjectEndMarker}
		assert.Equal(t, want, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(0)
		n, err := a.encodeObject(buf, AmfObject{"a": true})
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(1)
		n, err := a.encodeObject(buf, AmfObject{"a": true})
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{amf0ObjectMarker}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(4)
		n, err := a.encodeObject(buf, AmfObject{"a": true})
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		want := []byte{amf0ObjectMarker,
			0x00, 0x01, 'a'}
		assert.Equal(t, want, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(6)
		n, err := a.encodeObject(buf, AmfObject{"a": true})
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		want := []byte{amf0ObjectMarker,
			0x00, 0x01, 'a',
			amf0BooleanMarker, amf0BooleanTrue}
		assert.Equal(t, want, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(8)
		n, err := a.encodeObject(buf, AmfObject{"a": true})
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		want := []byte{amf0ObjectMarker,
			0x00, 0x01, 'a',
			amf0BooleanMarker, amf0BooleanTrue,
			0x00, 0x00}
		assert.Equal(t, want, buf.Bytes())
	}
}

func Test_amf0_encodeString(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		n, err := a.encodeString(buf, "abc", true)
		assert.Equal(t, 6, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0StringMarker, 0x00, 0x03, 'a', 'b', 'c'}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(0)
		n, err := a.encodeString(buf, "abc", true)
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(1)
		n, err := a.encodeString(buf, "abc", true)
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{amf0StringMarker}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(3)
		n, err := a.encodeString(buf, "abc", true)
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{amf0StringMarker, 0x00, 0x03}, buf.Bytes())
	}
}

func Test_amf0_encodeLongString(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		n, err := a.encodeLongString(buf, "abc")
		assert.Equal(t, 8, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0LongStringMarker, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(0)
		n, err := a.encodeLongString(buf, "abc")
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(1)
		n, err := a.encodeLongString(buf, "abc")
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{amf0LongStringMarker}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(5)
		n, err := a.encodeLongString(buf, "abc")
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, []byte{amf0LongStringMarker, 0x00, 0x00, 0x00, 0x03}, buf.Bytes())
	}
}

func Test_amf0_encode(t *testing.T) {
	a, _ := newAmf0()
	{
		buf := &bytes.Buffer{}
		var v interface{}
		n, err := a.encode(buf, v)
		assert.Equal(t, 1, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0NullMarker}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		var v bool
		n, err := a.encode(buf, v)
		assert.Equal(t, 2, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0BooleanMarker, amf0BooleanFalse}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		var v int8
		n, err := a.encode(buf, v)
		assert.Equal(t, 9, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0NumberMarker, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		var v uint8
		n, err := a.encode(buf, v)
		assert.Equal(t, 9, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0NumberMarker, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		var v float32
		n, err := a.encode(buf, v)
		assert.Equal(t, 9, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0NumberMarker, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		var v = AmfArray{true}
		n, err := a.encode(buf, v)
		assert.Equal(t, 7, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0StrictArrayMarker, 0x00, 0x00, 0x00, 0x01, amf0BooleanMarker, amf0BooleanTrue}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		var v = AmfObject{"a": true}
		n, err := a.encode(buf, v)
		assert.Equal(t, 9, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0ObjectMarker, 0x00, 0x01, 'a', amf0BooleanMarker, amf0BooleanTrue, 0x00, 0x00, amf0ObjectEndMarker}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		var v = map[int]interface{}{}
		n, err := a.encode(buf, v)
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Nil(t, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		var v = "abc"
		n, err := a.encode(buf, v)
		assert.Equal(t, 6, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{amf0StringMarker, 0x00, 0x03, 'a', 'b', 'c'}, buf.Bytes())
	}
	{
		buf := &bytes.Buffer{}
		input := &bytes.Buffer{}
		for i := 0; i < 1e6; i++ {
			input.WriteByte('a')
		}
		n, err := a.encode(buf, input.String())
		assert.Equal(t, 1000005, n)
		assert.NoError(t, err)
		s := buf.Bytes()
		assert.Equal(t, uint8(amf0LongStringMarker), s[0])
		assert.Equal(t, uint8(0x00), s[1])
		assert.Equal(t, uint8(0x0f), s[2])
		assert.Equal(t, uint8(0x42), s[3])
		assert.Equal(t, uint8(0x40), s[4])
		for i := 5; i < 1e6+5; i++ {
			if s[i] != uint8('a') {
				assert.Fail(t, "s[i]!='a'")
				break
			}
		}
	}
	{
		buf := &bytes.Buffer{}
		var v chan int
		n, err := a.encode(buf, v)
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Nil(t, buf.Bytes())
	}
}
