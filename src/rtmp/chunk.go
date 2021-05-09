package rtmp

import (
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zhyoulun/gls/src/utils"
	"io"
)

type chunkBasicHeader struct {
	chunkStreamID uint32
	fmt           Fmt
}

func newChunkBasicHeader() (*chunkBasicHeader, error) {
	return &chunkBasicHeader{}, nil
}

func newChunkBasicHeader2(chunkStreamID uint32, fmt Fmt) (*chunkBasicHeader, error) {
	return &chunkBasicHeader{
		chunkStreamID: chunkStreamID,
		fmt:           fmt,
	}, nil
}

func (cbh *chunkBasicHeader) Read(r io.Reader) error {
	b, err := utils.ReadByte(r)
	if err != nil {
		return errors.Wrap(err, "utils read byte")
	}
	cbh.fmt = Fmt(b >> 6)
	tmpChunkStreamID := b & 0x3f
	if tmpChunkStreamID == 0 { //2B, chunkStreamID: 64+[0,255]
		if b, err := utils.ReadByte(r); err != nil {
			return err
		} else {
			cbh.chunkStreamID = uint32(b) + 64
		}
	} else if tmpChunkStreamID == 1 { //3B,chunkStreamID: 64+[0,65535]
		var num uint16
		if err := binary.Read(r, binary.LittleEndian, &num); err != nil {
			return err
		} else {
			cbh.chunkStreamID = uint32(num) + 64
		}
	} else { //1B, chunkStreamID:[2,63]
		cbh.chunkStreamID = uint32(tmpChunkStreamID)
	}
	return nil
}

func (cbh *chunkBasicHeader) Write(w io.Writer) error {
	h := uint8(cbh.fmt) << 6
	switch {
	case cbh.chunkStreamID < 64:
		h |= uint8(cbh.chunkStreamID)
		if err := utils.WriteByte(w, h); err != nil {
			return err
		}
	case cbh.chunkStreamID-64 < 256:
		h |= 0
		if err := utils.WriteByte(w, h); err != nil {
			return err
		}
		if err := utils.WriteByte(w, uint8(cbh.chunkStreamID-64)); err != nil {
			return err
		}
	case cbh.chunkStreamID-64 < 65536:
		h |= 1
		if err := utils.WriteByte(w, h); err != nil {
			return err
		}
		num := uint16(cbh.chunkStreamID - 64)
		if err := binary.Write(w, binary.LittleEndian, &num); err != nil {
			return err
		}
	}
	return nil
}

func (cbh *chunkBasicHeader) String() string {
	return fmt.Sprintf("format: %d, chunk stream ID: %d", cbh.fmt, cbh.chunkStreamID)
}
