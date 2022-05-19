package qoi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
)

var ErrParseHeader = errors.New("failed to parse QOI header")

func Decode(reader io.Reader) (image.Image, error) {
	magic := make([]byte, 4)
	err := binary.Read(reader, binary.BigEndian, magic)
	if err != nil {
		return nil, err
	}

	correctMagic := []byte{'q', 'o', 'i', 'f'}
	if string(magic) != string(correctMagic) {
		return nil, fmt.Errorf("bad magic bytes: %w", ErrParseHeader)
	}

	var width uint32
	err = binary.Read(reader, binary.BigEndian, &width)
	if err != nil {
		return nil, err
	}

	var height uint32
	err = binary.Read(reader, binary.BigEndian, &height)
	if err != nil {
		return nil, err
	}

	var channels uint8
	err = binary.Read(reader, binary.BigEndian, &channels)
	if err != nil {
		return nil, err
	}
	if channels != ChannelRGB && channels != ChannelRGBA {
		return nil, fmt.Errorf("bad channels %v: %w", channels, ErrParseHeader)
	}

	var colorSpace uint8
	err = binary.Read(reader, binary.BigEndian, &colorSpace)
	if err != nil {
		return nil, err
	}
	if colorSpace != ColorSpaceSRGB && colorSpace != ColorSpaceLinear {
		return nil, fmt.Errorf("bad color space %v: %w", colorSpace, ErrParseHeader)
	}

	return image.NewNRGBA(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(width), int(height)},
	}), nil
}
