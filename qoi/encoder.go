package qoi

import (
	"encoding/binary"
	"image"
	"image/color"
	"io"
)

func Encode(w io.Writer, m image.Image) error {
	e := encoder{
		binWriter: binaryWriterErr{writer: w},
		image:     m,
		prev:      rgba{0, 0, 0, 255},
	}

	e.writeHeader()
	if e.binWriter.err != nil {
		return e.binWriter.err
	}

	for y := 0; y < m.Bounds().Dy(); y++ {
		for x := 0; x < m.Bounds().Dx(); x++ {
			e.writeChunk(x, y)
			if e.binWriter.err != nil {
				return e.binWriter.err
			}
		}
	}

	if e.runLength > 0 {
		e.writeRunChunk()
	}

	e.writeEndMarker()
	if e.binWriter.err != nil {
		return e.binWriter.err
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

func newRGBA(c color.Color) rgba {
	c = color.NRGBAModel.Convert(c)
	nrgba, ok := c.(color.NRGBA)
	if !ok {
		panic("couldn't convert to NRGBA")
	}

	return rgba{
		r: byte(nrgba.R),
		g: byte(nrgba.G),
		b: byte(nrgba.B),
		a: byte(nrgba.A),
	}
}

func hasAlpha(m image.Image) bool {
	rect := m.Bounds()
	for x := 0; x < rect.Dx(); x++ {
		for y := 0; y < rect.Dy(); y++ {
			_, _, _, a := m.At(x, y).RGBA()
			if byte(a) != byte(255) {
				return true
			}
		}
	}
	return false
}

type encoder struct {
	binWriter binaryWriterErr
	image     image.Image
	cache     [64]rgba
	prev      rgba
	runLength byte
}

func (e *encoder) writeHeader() {
	e.binWriter.write([]byte("qoif"))

	rect := e.image.Bounds()
	width := uint32(rect.Dx())
	e.binWriter.write(width)

	height := uint32(rect.Dy())
	e.binWriter.write(height)

	channel := ChannelsRGB
	if hasAlpha(e.image) {
		channel = ChannelsRGBA
	}
	e.binWriter.write(channel)

	e.binWriter.write(ColorSpaceSRGB)
}

func (e *encoder) isNewRun(next rgba) bool {
	return e.runLength == 0 && e.prev == next
}

func (e *encoder) canLengthenRun(next rgba) bool {
	return e.runLength > 0 && e.prev == next && e.runLength <= 61
}

func diff(prev rgba, next rgba) (byte, byte, byte) {
	dr := next.r - prev.r + 2
	dg := next.g - prev.g + 2
	db := next.b - prev.b + 2
	return dr, dg, db
}

func diffLuma(prev rgba, next rgba) (dg byte, drdg byte, dbdg byte) {
	dg = next.g - prev.g + 32
	drdg = (next.r - prev.r) - (next.g - prev.g) + 8
	dbdg = (next.b - prev.b) - (next.g - prev.g) + 8
	return
}

func isSmallDiff(diff byte) bool {
	return diff <= 3
}

func isSmallLumaDiff(dg, drdg, dbdg byte) bool {
	return dg <= 63 && drdg <= 15 && dbdg <= 15
}

func (e *encoder) writeChunk(x, y int) {
	pixel := newRGBA(e.image.At(x, y))
	index := calculateIndex(pixel)
	cachePixel := e.cache[index]

	switch {
	case e.isNewRun(pixel) || e.canLengthenRun(pixel):
		e.cache[index] = pixel
		e.runLength++

	case e.runLength > 0:
		e.writeRunChunk()
		if e.binWriter.err != nil {
			return
		}

		e.runLength = 0
		e.writeChunk(x, y)
		return

	case pixel == cachePixel:
		e.writeIndexChunk(index)

	case e.prev.a == pixel.a:
		e.cache[index] = pixel

		dr, dg, db := diff(e.prev, pixel)
		if isSmallDiff(dr) && isSmallDiff(dg) && isSmallDiff(db) {
			e.writeDiffChunk(dr, dg, db)
			break
		}

		dgLuma, drdg, dbdg := diffLuma(e.prev, pixel)
		if isSmallLumaDiff(dgLuma, drdg, dbdg) {
			e.writeLumaChunk(dgLuma, drdg, dbdg)
			break
		}

		e.writeRGBChunk(pixel)

	default:
		e.cache[index] = pixel
		e.writeRGBAChunk(pixel)
	}

	e.prev = pixel
}

func (e *encoder) writeRGBChunk(pixel rgba) {
	e.binWriter.write(byte(0b11111110))
	e.binWriter.write(pixel.r)
	e.binWriter.write(pixel.g)
	e.binWriter.write(pixel.b)
}

func (e *encoder) writeRGBAChunk(pixel rgba) {
	e.binWriter.write(byte(0b11111111))
	e.binWriter.write(pixel.r)
	e.binWriter.write(pixel.g)
	e.binWriter.write(pixel.b)
	e.binWriter.write(pixel.a)
}

func (e *encoder) writeIndexChunk(index int) {
	e.binWriter.write(byte(index))
}

func (e *encoder) writeDiffChunk(dr byte, dg byte, db byte) {
	chunk := byte(0b01000000)
	chunk |= dr << 4
	chunk |= dg << 2
	chunk |= db
	e.binWriter.write(chunk)
}

func (e *encoder) writeLumaChunk(dg byte, drdg byte, dbdg byte) {
	first := byte(0b10000000)
	first |= dg
	e.binWriter.write(first)
	second := byte(0)
	second |= drdg << 4
	second |= dbdg
	e.binWriter.write(second)
}

func (e *encoder) writeRunChunk() {
	chunk := byte(0b11000000)
	chunk |= e.runLength - 1
	e.binWriter.write(chunk)
}

func calculateIndex(color rgba) int {
	return int((color.r*3 + color.g*5 + color.b*7 + color.a*11) % 64)
}

func (e *encoder) writeEndMarker() {
	e.binWriter.write([]byte{0, 0, 0, 0, 0, 0, 0, 1})
}