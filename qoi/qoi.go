package qoi

import (
	"encoding/binary"
	"image"
	"image/color"
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
	for y := 0; y < m.Bounds().Dy(); y++ {
		for x := 0; x < m.Bounds().Dx(); x++ {
			err = e.writeChunk(x, y)
			if err != nil {
				return err
			}
		}
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

type rgba struct {
	r byte
	g byte
	b byte
	a byte
}

func newRGBA(color color.Color) rgba {
	r, g, b, a := color.RGBA()
	return rgba{
		r: byte(r),
		g: byte(g),
		b: byte(b),
		a: byte(a),
	}
}

type encoder struct {
	writer io.Writer
	image  image.Image
	cache  [64]rgba
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

func (e *encoder) writeChunk(x, y int) error {
	var binWriter binaryWriterErr
	previousAlpha := byte(255)
	pixel := newRGBA(e.image.At(x, y))
	index := calculateIndex(pixel)
	cachePixel := e.cache[index]
	if pixel == cachePixel {
		binWriter.write(e.writer, binary.BigEndian, byte(index))
	} else if previousAlpha == pixel.a {
		binWriter.write(e.writer, binary.BigEndian, byte(0b11111110))
		binWriter.write(e.writer, binary.BigEndian, pixel.r)
		binWriter.write(e.writer, binary.BigEndian, pixel.g)
		binWriter.write(e.writer, binary.BigEndian, pixel.b)
		e.cache[index] = pixel
	} else {
		binWriter.write(e.writer, binary.BigEndian, byte(0b11111111))
		binWriter.write(e.writer, binary.BigEndian, pixel.r)
		binWriter.write(e.writer, binary.BigEndian, pixel.g)
		binWriter.write(e.writer, binary.BigEndian, pixel.b)
		binWriter.write(e.writer, binary.BigEndian, pixel.a)
		e.cache[index] = pixel
	}
	return binWriter.err
}

func calculateIndex(color rgba) int {
	return int((color.r*3 + color.g*5 + color.b*7 + color.a*11) % 64)
}

func (e *encoder) writeEndMarker() error {
	_, err := e.writer.Write([]byte{0, 0, 0, 0, 0, 0, 0, 1})
	return err
}
