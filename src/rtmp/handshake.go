package rtmp

import (
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/zhyoulun/gls/src/utils"
	"io"
	"math/rand"
	"time"
)

type handshake struct {
	c0       byte
	c1time   uint32
	c1random [1528]byte
	c2time   uint32
	c2time2  uint32
	c2random [1528]byte
	s1random [1528]byte
}

func (h *handshake) String() string {
	return fmt.Sprintf("c0: %+v, c1time: %+v, c1random: %+v, c2time: %+v, c2time2: %+v, c2random: %+v, s1random: %+v",
		h.c0, h.c1time, h.c1random, h.c2time, h.c2time2, h.c2random, h.s1random)
}

func newHandshake() (*handshake, error) {
	rand.Seed(time.Now().UnixNano())
	return &handshake{}, nil
}

func (h *handshake) Do(rw io.ReadWriter) error {
	if err := h.readC0(rw); err != nil {
		return err
	}

	//the server MUST wait until C0 has been received before sending S0 and S1,
	//and MAY wait until after C1 as well
	if err := h.writeS0(rw); err != nil {
		return err
	}
	if err := h.writeS1(rw); err != nil {
		return err
	}

	if err := h.readC1(rw); err != nil {
		return err
	}

	//the server MUST wait until C1 has been received before sending S2
	if err := h.writeS2(rw); err != nil {
		return err
	}

	//the send MUST wait until C2 has been received before sending any other data
	if err := h.readC2(rw); err != nil {
		return err
	}
	log.Tracef("handshake: %s", h)
	return nil
}

func (h *handshake) readC0(r io.Reader) error {
	if b, err := utils.ReadByte(r); err != nil {
		return err
	} else {
		//the version defined by this specification is 3.
		//Value 0-2 are deprecated values used by earlier proprietary products;
		//4-31 are reserved for future implementations;
		//and 32-255 are not allowed(to allow distinguishing RTMP from text-based protocols, which always start with a printable character)
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
		//this field MUST be all 0s
		if zero != 0 {
			log.Warnf("read c1 zero, want 0, got: %d", zero)
			//todo 遇到兼容问题了，降低要求，badcase: ffmpeg version 4.4
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
	//this field MUST contain the random data field sent by the peer in S1 (for C2) or S2 (for C1).
	if random, err := utils.ReadBytes(r, 1528); err != nil {
		return err
	} else {
		copy(h.c2random[:], random)
	}
	if h.s1random != h.c2random {
		return errors.Errorf("read c2 random, want c2random=s1random, bot not")
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
	//this field can contain any arbitrary values. Since each endpoint has to distinguish
	//between the response to the handshake it has initiated
	//and the handshake initiated by its peer,
	//this data SHOULD（不强制要求） send something sufficiently random.
	//But there is no need for cryptographically-secure randomness, or even dynamic values
	var random [1528]byte
	_, _ = rand.Read(random[:])
	h.s1random = random
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
