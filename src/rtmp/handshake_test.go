package rtmp

import (
	"bou.ke/monkey"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/zhyoulun/gls/src/utils"
	"math/rand"
	"os"
	"testing"
)

func Test_handshake_readC0(t *testing.T) {
	{
		r := bytes.NewReader([]byte{rtmpVersion})
		h, _ := newHandshake()
		err := h.readC0(r)
		assert.NoError(t, err)
		assert.Equal(t, byte(rtmpVersion), h.c0)
		fmt.Fprintf(os.Stdout, "%s", h)
	}
	{
		r := bytes.NewReader([]byte{1})
		h, _ := newHandshake()
		err := h.readC0(r)
		assert.Error(t, err)
	}
	{
		r := bytes.NewReader([]byte{})
		h, _ := newHandshake()
		err := h.readC0(r)
		assert.Error(t, err)
	}
}

func Test_handshake_readC1(t *testing.T) {
	{
		src := []byte{0x01, 0x02, 0x03, 0x04,
			0x00, 0x00, 0x00, 0x00}
		rand := [1528]byte{0x01}
		src = append(src, rand[:]...)
		r := bytes.NewReader(src)
		h, _ := newHandshake()
		err := h.readC1(r)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x01020304), h.c1time)
		assert.Equal(t, [1528]byte{1}, h.c1random)
	}
	{
		src := []byte(``)
		r := bytes.NewReader(src)
		h, _ := newHandshake()
		err := h.readC1(r)
		assert.Error(t, err)
	}
	{
		src := []byte{0x01, 0x02, 0x03, 0x04,
			0x00, 0x00, 0x00, 0x01}
		r := bytes.NewReader(src)
		h, _ := newHandshake()
		err := h.readC1(r)
		assert.Error(t, err)
	}
	{
		src := []byte{0x01, 0x02, 0x03, 0x04}
		r := bytes.NewReader(src)
		h, _ := newHandshake()
		err := h.readC1(r)
		assert.Error(t, err)
	}
	{
		src := []byte{0x01, 0x02, 0x03, 0x04,
			0x00, 0x00, 0x00, 0x00}
		r := bytes.NewReader(src)
		h, _ := newHandshake()
		err := h.readC1(r)
		assert.Error(t, err)
	}
}

func Test_handshake_readC2(t *testing.T) {
	{
		src := []byte{0x01, 0x02, 0x03, 0x04,
			0x05, 0x06, 0x07, 0x08}
		randBytes := [1528]byte{0x01}
		src = append(src, randBytes[:]...)
		r := bytes.NewReader(src)
		h, _ := newHandshake()
		h.s1random = randBytes
		err := h.readC2(r)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x01020304), h.c2time)
		assert.Equal(t, uint32(0x05060708), h.c2time2)
		assert.Equal(t, [1528]byte{1}, h.c2random)
	}
	{
		src := []byte(``)
		r := bytes.NewReader(src)
		h, _ := newHandshake()
		err := h.readC2(r)
		assert.Error(t, err)
	}
	{
		src := []byte{0x01, 0x02, 0x03, 0x04}
		r := bytes.NewReader(src)
		h, _ := newHandshake()
		err := h.readC2(r)
		assert.Error(t, err)
	}
	{
		src := []byte{0x01, 0x02, 0x03, 0x04,
			0x05, 0x06, 0x07, 0x08}
		r := bytes.NewReader(src)
		h, _ := newHandshake()
		err := h.readC2(r)
		assert.Error(t, err)
	}
	{
		src := []byte{0x01, 0x02, 0x03, 0x04,
			0x05, 0x06, 0x07, 0x08}
		randBytes := [1528]byte{0x01}
		src = append(src, randBytes[:]...)
		r := bytes.NewReader(src)
		h, _ := newHandshake()
		h.s1random = [1528]byte{0x02}
		err := h.readC2(r)
		assert.Error(t, err)
	}
}

func Test_handshake_writeS0(t *testing.T) {
	{
		buf := &bytes.Buffer{}
		h, _ := newHandshake()
		err := h.writeS0(buf)
		assert.NoError(t, err)
		assert.Equal(t, []byte{rtmpVersion}, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(0)
		h, _ := newHandshake()
		err := h.writeS0(buf)
		assert.Error(t, err)
	}
}

func Test_handshake_writeS1(t *testing.T) {
	monkey.Patch(rand.Read, func(p []byte) (n int, err error) {
		p[0] = 0x01
		return len(p), nil
	})
	{
		buf := &bytes.Buffer{}
		h, _ := newHandshake()
		err := h.writeS1(buf)
		assert.NoError(t, err)
		expect := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		random := [1528]byte{0x01}
		expect = append(expect, random[:]...)
		assert.Equal(t, expect, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(193)
		h, _ := newHandshake()
		err := h.writeS1(buf)
		assert.Error(t, err)
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(7)
		h, _ := newHandshake()
		err := h.writeS1(buf)
		assert.Error(t, err)
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(3)
		h, _ := newHandshake()
		err := h.writeS1(buf)
		assert.Error(t, err)
	}
}

func Test_handshake_writeS2(t *testing.T) {
	{
		buf := &bytes.Buffer{}
		h, _ := newHandshake()
		h.c1time = 0x01020304
		err := h.writeS2(buf)
		assert.NoError(t, err)
		expect := []byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}
		random := [1528]byte{}
		expect = append(expect, random[:]...)
		assert.Equal(t, expect, buf.Bytes())
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(193)
		h, _ := newHandshake()
		err := h.writeS2(buf)
		assert.Error(t, err)
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(7)
		h, _ := newHandshake()
		err := h.writeS2(buf)
		assert.Error(t, err)
	}
	{
		buf, _ := utils.NewBufferWithMaxCapacity(3)
		h, _ := newHandshake()
		err := h.writeS2(buf)
		assert.Error(t, err)
	}
}
