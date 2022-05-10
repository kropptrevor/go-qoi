package qoi

import (
	"encoding/binary"
	"image"
	"io"
)

const (
	ChannelRGBA    uint8 = 4
	ColorSpaceSRGB uint8 = 0
)

func Encode(w io.Writer, m image.Image) error {
	binary.Write(w, binary.BigEndian, []byte("qoif"))
	rect := m.Bounds()
	width := uint32(rect.Dx())
	binary.Write(w, binary.BigEndian, width)
	height := uint32(rect.Dy())
	binary.Write(w, binary.BigEndian, height)
	binary.Write(w, binary.BigEndian, ChannelRGBA)
	return nil
}
