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
	e := encoder{
		writer: w,
		image:  m,
	}
	e.writeHeader()
	e.writeEndMarker()
	return nil
}

type encoder struct {
	writer io.Writer
	image  image.Image
}

func (e encoder) writeHeader() {
	binary.Write(e.writer, binary.BigEndian, []byte("qoif"))
	rect := e.image.Bounds()
	width := uint32(rect.Dx())
	binary.Write(e.writer, binary.BigEndian, width)
	height := uint32(rect.Dy())
	binary.Write(e.writer, binary.BigEndian, height)
	binary.Write(e.writer, binary.BigEndian, ChannelRGBA)
	binary.Write(e.writer, binary.BigEndian, ColorSpaceSRGB)
}

func (e encoder) writeEndMarker() {
	binary.Write(e.writer, binary.BigEndian, byte(0))
	binary.Write(e.writer, binary.BigEndian, byte(0))
	binary.Write(e.writer, binary.BigEndian, byte(0))
	binary.Write(e.writer, binary.BigEndian, byte(0))
	binary.Write(e.writer, binary.BigEndian, byte(0))
	binary.Write(e.writer, binary.BigEndian, byte(0))
	binary.Write(e.writer, binary.BigEndian, byte(0))
	binary.Write(e.writer, binary.BigEndian, byte(1))
}
