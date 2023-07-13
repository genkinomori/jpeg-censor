package core

import (
	"errors"
)

var MAGIC = []byte("GENKINOMORI")

type Header []uint8

func NewHeader() *Header {
	ret := make(Header, 0, 512) // use a slightly large capacity
	return &ret
}

func (h *Header) PushBit(bit uint8) {
	*h = append(*h, bit&0x1)
}

func (h *Header) PushByte(b uint8) {
	s := make([]uint8, 8)
	for i := 0; i < 8; i++ {
		s[i] = b & 0x1
		b >>= 1
	}
	*h = append(*h, s...)
}

func (h *Header) Len() int {
	return len(*h)
}

func (h *Header) Get(i int) uint8 {
	return (*h)[i]
}

func (h *Header) Pop() uint8 {
	if h.Len() == 0 {
		panic("unexpected EOF")
	}
	ret := (*h)[0]
	*h = (*h)[1:]
	return ret
}

func (h *Header) PushUint16(v uint16) {
	s := make([]uint8, 16)
	for i := 0; i < 16; i++ {
		s[i] = uint8(v & 0x1)
		v >>= 1
	}
	*h = append(*h, s...)
}

func (h *Header) PushUint32(v uint32) {
	s := make([]uint8, 32)
	for i := 0; i < 32; i++ {
		s[i] = uint8(v & 0x1)
		v >>= 1
	}
	*h = append(*h, s...)
}

func (h *Header) PushUint16V(v uint16) {
	if v <= 0xFE {
		h.PushByte(uint8(v))
	} else {
		h.PushByte(0xFF)
		h.PushUint16(v)
	}
}

func (h *Header) PushUint32V(v uint32) {
	if v <= 0xFFFE {
		h.PushUint16V(uint16(v))
	} else {
		h.PushUint16V(0xFFFF)
		h.PushUint16(uint16(v & 0xFFFF))
		v >>= 16
		h.PushUint16(uint16(v))
	}
}

func (h *Header) PopByte() uint8 {
	ret := uint8(0)
	for i := 0; i < 8; i++ {
		ret |= uint8(h.Pop()) << i
	}
	return ret
}

func (h *Header) PopUint16() uint16 {
	ret := uint16(0)
	for i := 0; i < 16; i++ {
		ret |= uint16(h.Pop()) << i
	}
	return ret
}

func (h *Header) PopUint32() uint32 {
	ret := uint32(0)
	for i := 0; i < 32; i++ {
		ret |= uint32(h.Pop()) << i
	}
	return ret
}

func (h *Header) PopUint16V() uint16 {
	b := h.PopByte()
	if b <= 0xFE {
		return uint16(b)
	}
	return h.PopUint16()
}

func (h *Header) PopUint32V() uint32 {
	v := h.PopUint16V()
	if v <= 0xFFFE {
		return uint32(v)
	}
	ll := uint32(h.PopUint16())
	hh := uint32(h.PopUint16())
	return (hh << 16) | ll
}

func (h *Header) Concat(r *Header) {
	*h = append(*h, (*r)...)
}

func MakeHeader(cr *CensorResult) *Header {
	h := NewHeader()

	height := cr.CensoredImage.Bounds().Dy()
	h.PushUint32V(uint32(height))

	blocks := len(cr.Blocks)
	h.PushUint32V(uint32(blocks))

	var (
		ones    uint32 = 0
		current uint32 = 0
	)
	for _, b := range cr.Blocks {
		k := uint32(b.Index)
		d := k - current
		if d == 1 {
			ones++
		} else {
			h.PushUint32V(ones)
			h.PushUint32V(d)
			ones = 0
		}
		current = k
	}
	h.PushUint32V(ones)
	h.PushUint32V(0)

	meta := NewHeader()
	for _, b := range MAGIC {
		meta.PushByte(b)
	}
	ll := uint32(h.Len())
	meta.PushUint32(ll)
	meta.PushUint32(ll)
	meta.Concat(h)

	return meta
}

type ParsedHeader struct {
	Height     uint32
	Blocks     uint32
	BlockIndex []uint32
}

var (
	ErrMagicNotMatch       = errors.New("magic not match")
	ErrLengthNotDuplicated = errors.New("length not duplicated")
)

func ParseMagicAndLength(h *Header) (uint32, error) {
	for _, b := range MAGIC {
		bb := h.PopByte()
		if b != bb {
			return 0, ErrMagicNotMatch
		}
	}
	v1 := h.PopUint32()
	v2 := h.PopUint32()
	if v1 != v2 {
		return 0, ErrLengthNotDuplicated
	}
	return v1, nil
}

func ParseHeader(h *Header) *ParsedHeader {
	ret := &ParsedHeader{}
	ret.Height = h.PopUint32V()
	ret.Blocks = h.PopUint32V()
	blockIndex := make([]uint32, 0, ret.Blocks)
	p := uint32(0)
	for {
		ones := h.PopUint32V()
		d := h.PopUint32V()
		for i := uint32(0); i < ones; i++ {
			p++
			blockIndex = append(blockIndex, p)
		}
		if d == 0 {
			break
		}
		p += d
		blockIndex = append(blockIndex, p)
	}
	if len(blockIndex) != int(ret.Blocks) {
		panic("block count not consistent")
	}
	ret.BlockIndex = blockIndex
	return ret
}
