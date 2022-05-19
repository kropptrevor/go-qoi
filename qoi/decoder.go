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
	binary.Read(d.input, binary.BigEndian, &b)
	if b == TagRGB {
		bs := [3]byte{}
		err := binary.Read(d.input, binary.BigEndian, &bs)
		if err != nil {
			return err
		}

		pixel := rgba{bs[0], bs[1], bs[2], 255}
		index := pixel.index()
		d.cache[index] = pixel
		d.img.SetRGBA(x, y, color.RGBA(pixel))
	} else if b == TagRGBA {
		bs := [4]byte{}
		err := binary.Read(d.input, binary.BigEndian, &bs)
		if err != nil {
			return err
		}

		pixel := rgba{bs[0], bs[1], bs[2], bs[3]}
		index := pixel.index()
		d.cache[index] = pixel
		d.img.SetRGBA(x, y, color.RGBA(pixel))
	} else if b&TagMask == TagIndex {
		index := b & ^TagMask
		pixel := d.cache[index]
		d.img.SetRGBA(x, y, color.RGBA(pixel))
	}
	return nil
}
