package amf

const (
	amf0NumberMarker        = 0x00
	amf0BooleanMarker       = 0x01
	amf0StringMarker        = 0x02
	amf0ObjectMarker        = 0x03
	amf0MovieclipMarker     = 0x04
	amf0NullMarker          = 0x05
	amf0UndefinedMarker     = 0x06
	amf0ReferenceMarker     = 0x07
	amf0EcmaArrayMarker     = 0x08
	amf0ObjectEndMarker     = 0x09
	amf0StrictArrayMarker   = 0x0a
	amf0DateMarker          = 0x0b
	amf0LongStringMarker    = 0x0c
	amf0UnsupportedMarker   = 0x0d
	amf0RecordsetMarker     = 0x0e
	amf0XmlDocumentMarker   = 0x0f
	amf0TypedObjectMarker   = 0x10
	amf0AvmplusObjectMarker = 0x11
)

const (
	amf0StringMax = 65535
)

const (
	amf0BooleanTrue  = 0x01
	amf0BooleanFalse = 0x00
)

type AmfVersion int

const (
	Amf0 AmfVersion = 0x00
	Amf3 AmfVersion = 0x03
)
