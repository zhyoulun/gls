package rtmp

import (
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zhyoulun/gls/src/utils"
	"io"
)

type handshake struct {
	c0       byte
	c1time   uint32
	c1random [1528]byte
	c2time   uint32
	c2time2  uint32
	c2random [1528]byte
}

func (h *handshake) String() string {
	return fmt.Sprintf("c0: %+v, c1time: %+v, c1random: %+v, c2time: %+v, c2time2: %+v, c2random: %+v",
		h.c0, h.c1time, h.c1random, h.c2time, h.c2time2, h.c2random)
}

func newHandshake() (*handshake, error) {
	return &handshake{}, nil
}

func (h *handshake) readC0(r io.Reader) error {
	if b, err := utils.ReadByte(r); err != nil {
		return err
	} else {
		if b != rtmpVersion {
			return fmt.Errorf("read c0, want: %d, got: %d", rtmpVersion, b)
		}
		h.c0 = b
	}
	return nil
}

func (h *handshake) readC1(r io.Reader) error {
	//read time
	var timestamp uint32
	if err := binary.Read(r, binary.BigEndian, &timestamp); err != nil {
		return err
	} else {
		h.c1time = timestamp
	}

	//read zero
	var zero uint32
	if err := binary.Read(r, binary.BigEndian, &zero); err != nil {
		return err
	} else {
		if zero != 0 {
			log.Warnf("read c1 zero, want 0, got: %d", zero)
			//return fmt.Errorf("read c1 zero, want 0, got: %d", zero)
		}
	}

	//read random
	if random, err := utils.ReadBytes(r, 1528); err != nil {
		return err
	} else {
		copy(h.c1random[:], random)
	}
	return nil
}

func (h *handshake) readC2(r io.Reader) error {
	//read time
	var timestamp uint32
	if err := binary.Read(r, binary.BigEndian, &timestamp); err != nil {
		return err
	} else {
		h.c2time = timestamp
	}

	//read time2
	var timestamp2 uint32
	if err := binary.Read(r, binary.BigEndian, &timestamp2); err != nil {
		return err
	} else {
		h.c2time2 = timestamp2
	}

	//read random
	if random, err := utils.ReadBytes(r, 1528); err != nil {
		return err
	} else {
		copy(h.c2random[:], random)
	}
	return nil
}

func (h *handshake) writeS0(w io.Writer) error {
	if err := utils.WriteByte(w, rtmpVersion); err != nil {
		return err
	}
	return nil
}

func (h *handshake) writeS1(w io.Writer) error {
	var timestamp uint32 = 0
	if err := binary.Write(w, binary.BigEndian, &timestamp); err != nil {
		return err
	}
	var zero uint32 = 0
	if err := binary.Write(w, binary.BigEndian, &zero); err != nil {
		return err
	}
	var random [1528]byte //todo 待优化，这里不应该全是0
	if err := utils.WriteBytes(w, random[:]); err != nil {
		return err
	}
	return nil
}

func (h *handshake) writeS2(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, &h.c1time); err != nil {
		return err
	}
	var zero uint32 = 0
	if err := binary.Write(w, binary.BigEndian, &zero); err != nil {
		return err
	}
	if err := utils.WriteBytes(w, h.c1random[:]); err != nil {
		return err
	}
	return nil
}
