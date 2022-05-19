package qoi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
)

var ErrParseHeader = errors.New("failed to parse QOI header")

var ErrParseEndMarker = errors.New("failed to parse QOI end marker")

func Decode(input io.Reader) (image.Image, error) {
	d := decoder{
		input: input,
		cache: [64]rgba{},
	}
	width, height, err := d.parseHeader()
	if err != nil {
		return nil, err
	}

	d.img = image.NewRGBA(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(width), int(height)},
	})

	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			d.parseChunk(x, y)
		}
	}

	err = d.parseEndMarker()
	if err != nil {
		return nil, err
	}

	return d.img, nil
}

type decoder struct {
	input io.Reader
	cache [64]rgba
	img   *image.RGBA
	prev  rgba
}

func (d *decoder) parseHeader() (width uint32, height uint32, err error) {
	magic := make([]byte, 4)
	err = binary.Read(d.input, binary.BigEndian, magic)
	if err != nil {
		return 0, 0, err
	}

	correctMagic := []byte{'q', 'o', 'i', 'f'}
	if string(magic) != string(correctMagic) {
		return 0, 0, fmt.Errorf("bad magic bytes: %w", ErrParseHeader)
	}

	err = binary.Read(d.input, binary.BigEndian, &width)
	if err != nil {
		return 0, 0, err
	}

	err = binary.Read(d.input, binary.BigEndian, &height)
	if err != nil {
		return 0, 0, err
	}

	var channels uint8
	err = binary.Read(d.input, binary.BigEndian, &channels)
	if err != nil {
		return 0, 0, err
	}
	if channels != ChannelsRGB && channels != ChannelsRGBA {
		return 0, 0, fmt.Errorf("bad channels %v: %w", channels, ErrParseHeader)
	}

	var colorSpace uint8
	err = binary.Read(d.input, binary.BigEndian, &colorSpace)
	if err != nil {
		return 0, 0, err
	}
	if colorSpace != ColorSpaceSRGB && colorSpace != ColorSpaceLinear {
		return 0, 0, fmt.Errorf("bad color space %v: %w", colorSpace, ErrParseHeader)
	}

	return
}

func (d *decoder) parseEndMarker() error {
	var endMarker uint64
	err := binary.Read(d.input, binary.BigEndian, &endMarker)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("missing end marker: %w", ErrParseEndMarker)
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return fmt.Errorf("partial end marker: %w", ErrParseEndMarker)
		}
		return err
	}
	if endMarker != 1 {
		return fmt.Errorf("bad end marker %v: %w", endMarker, ErrParseEndMarker)
	}

	return nil
}

func (d *decoder) parseChunk(x, y int) error {
	var b byte
	err := binary.Read(d.input, binary.BigEndian, &b)
	if err != nil {
		return err
	}
	var pixel rgba
	switch {
	case b == TagRGB:
		bs := [3]byte{}
		err = binary.Read(d.input, binary.BigEndian, &bs)
		if err != nil {
			return err
		}

		pixel = rgba{bs[0], bs[1], bs[2], 255}
		index := pixel.index()
		d.cache[index] = pixel

	case b == TagRGBA:
		bs := [4]byte{}
		err = binary.Read(d.input, binary.BigEndian, &bs)
		if err != nil {
			return err
		}

		pixel = rgba{bs[0], bs[1], bs[2], bs[3]}
		index := pixel.index()
		d.cache[index] = pixel

	case b&TagMask == TagIndex:
		index := b & ^TagMask
		pixel = d.cache[index]

	case b&TagMask == TagDiff:
		const bias = 2
		dr := (b&0b_11_00_00)>>4 - bias
		dg := (b&0b_00_11_00)>>2 - bias
		db := (b&0b_00_00_11)>>0 - bias

		pixel = d.prev
		pixel.R += dr
		pixel.G += dg
		pixel.B += db

	case b&TagMask == TagLuma:
		var b2 byte
		err = binary.Read(d.input, binary.BigEndian, &b2)
		if err != nil {
			return err
		}

		const gBias = 32
		const rbBias = 8
		dg := (b & ^TagMask)>>0 - gBias
		drdg := (b2&0b_1111_0000)>>4 - rbBias
		dbdg := (b2&0b_0000_1111)>>0 - rbBias

		pixel = d.prev
		pixel.R += (drdg + dg)
		pixel.G += dg
		pixel.B += (dbdg + dg)
	}
	d.img.SetRGBA(x, y, color.RGBA(pixel))
	d.prev = pixel
	return nil
}
