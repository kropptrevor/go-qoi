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
	err := e.writeHeader()
	if err != nil {
		return err
	}
	err = e.writeRGBA()
	if err != nil {
		return err
	}
	err = e.writeEndMarker()
	if err != nil {
		return err
	}
	return nil
}

type binaryWriterErr struct {
	err error
}

func (b *binaryWriterErr) write(w io.Writer, order binary.ByteOrder, data any) {
	if b == nil {
		return
	}
	err := binary.Write(w, order, data)
	if err != nil {
		b.err = err
	}
}

type encoder struct {
	writer io.Writer
	image  image.Image
}

func (e *encoder) writeHeader() error {
	var binWriter binaryWriterErr
	binWriter.write(e.writer, binary.BigEndian, []byte("qoif"))
	rect := e.image.Bounds()
	width := uint32(rect.Dx())
	binWriter.write(e.writer, binary.BigEndian, width)
	height := uint32(rect.Dy())
	binWriter.write(e.writer, binary.BigEndian, height)
	binWriter.write(e.writer, binary.BigEndian, ChannelRGBA)
	binWriter.write(e.writer, binary.BigEndian, ColorSpaceSRGB)
	return binWriter.err
}

func (e encoder) writeRGBA() error {
	var binWriter binaryWriterErr
	binWriter.write(e.writer, binary.BigEndian, byte(0b11111111))
	r, g, b, a := e.image.At(0, 0).RGBA()
	binWriter.write(e.writer, binary.BigEndian, byte(r))
	binWriter.write(e.writer, binary.BigEndian, byte(g))
	binWriter.write(e.writer, binary.BigEndian, byte(b))
	binWriter.write(e.writer, binary.BigEndian, byte(a))
	return binWriter.err
}

func (e encoder) writeEndMarker() error {
	var binWriter binaryWriterErr
	binWriter.write(e.writer, binary.BigEndian, byte(0))
	binWriter.write(e.writer, binary.BigEndian, byte(0))
	binWriter.write(e.writer, binary.BigEndian, byte(0))
	binWriter.write(e.writer, binary.BigEndian, byte(0))
	binWriter.write(e.writer, binary.BigEndian, byte(0))
	binWriter.write(e.writer, binary.BigEndian, byte(0))
	binWriter.write(e.writer, binary.BigEndian, byte(0))
	binWriter.write(e.writer, binary.BigEndian, byte(1))
	return binWriter.err
}
