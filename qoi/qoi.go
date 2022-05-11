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
		prev:   rgba{0, 0, 0, 255},
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
	writer io.Writer
	err    error
}

func (b *binaryWriterErr) write(data any) {
	if b == nil {
		return
	}
	err := binary.Write(b.writer, binary.BigEndian, data)
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
	prev   rgba
}

func (e *encoder) writeHeader() error {
	binWriter := binaryWriterErr{writer: e.writer}
	binWriter.write([]byte("qoif"))
	rect := e.image.Bounds()
	width := uint32(rect.Dx())
	binWriter.write(width)
	height := uint32(rect.Dy())
	binWriter.write(height)
	binWriter.write(ChannelRGBA)
	binWriter.write(ColorSpaceSRGB)
	return binWriter.err
}

func diff(prev rgba, next rgba) (int, int, int) {
	dr := int(next.r) - int(prev.r)
	dg := int(next.g) - int(prev.g)
	db := int(next.b) - int(prev.b)
	return dr, dg, db
}

func isSmallDiff(diff int) bool {
	return diff >= -2 && diff <= 1
}

func (e *encoder) writeChunk(x, y int) error {
	binWriter := binaryWriterErr{writer: e.writer}
	pixel := newRGBA(e.image.At(x, y))
	index := calculateIndex(pixel)
	cachePixel := e.cache[index]
	if pixel == cachePixel {
		e.writeIndexChunk(&binWriter, index)
	} else if e.prev.a == pixel.a {
		dr, dg, db := diff(e.prev, pixel)
		if isSmallDiff(dr) && isSmallDiff(dg) && isSmallDiff(db) {
			e.writeDiffChunk(&binWriter, dr, dg, db)
		} else {
			e.writeRGBChunk(&binWriter, pixel)
		}
		e.cache[index] = pixel
	} else {
		e.writeRGBAChunk(&binWriter, pixel)
		e.cache[index] = pixel
	}
	e.prev = pixel
	return binWriter.err
}

func (e *encoder) writeRGBChunk(binWriter *binaryWriterErr, pixel rgba) {
	binWriter.write(byte(0b11111110))
	binWriter.write(pixel.r)
	binWriter.write(pixel.g)
	binWriter.write(pixel.b)
}

func (e *encoder) writeRGBAChunk(binWriter *binaryWriterErr, pixel rgba) {
	binWriter.write(byte(0b11111111))
	binWriter.write(pixel.r)
	binWriter.write(pixel.g)
	binWriter.write(pixel.b)
	binWriter.write(pixel.a)
}

func (e *encoder) writeIndexChunk(binWriter *binaryWriterErr, index int) {
	binWriter.write(byte(index))
}

func (e *encoder) writeDiffChunk(binWriter *binaryWriterErr, dr int, dg int, db int) {
	chunk := byte(0b01000000)
	chunk |= byte(dr+2) << 4
	chunk |= byte(dg+2) << 2
	chunk |= byte(db + 2)
	binWriter.write(chunk)
}

func calculateIndex(color rgba) int {
	return int((color.r*3 + color.g*5 + color.b*7 + color.a*11) % 64)
}

func (e *encoder) writeEndMarker() error {
	_, err := e.writer.Write([]byte{0, 0, 0, 0, 0, 0, 0, 1})
	return err
}
