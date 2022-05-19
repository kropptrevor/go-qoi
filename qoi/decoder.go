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
	width, height, err := parseHeader(input)
	if err != nil {
		return nil, err
	}

	output := image.NewRGBA(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(width), int(height)},
	})

	size := width * height
	if size > 0 {
		parseChunk(input, output)
	}

	err = parseEndMarker(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func parseHeader(input io.Reader) (width uint32, height uint32, err error) {
	magic := make([]byte, 4)
	err = binary.Read(input, binary.BigEndian, magic)
	if err != nil {
		return 0, 0, err
	}

	correctMagic := []byte{'q', 'o', 'i', 'f'}
	if string(magic) != string(correctMagic) {
		return 0, 0, fmt.Errorf("bad magic bytes: %w", ErrParseHeader)
	}

	err = binary.Read(input, binary.BigEndian, &width)
	if err != nil {
		return 0, 0, err
	}

	err = binary.Read(input, binary.BigEndian, &height)
	if err != nil {
		return 0, 0, err
	}

	var channels uint8
	err = binary.Read(input, binary.BigEndian, &channels)
	if err != nil {
		return 0, 0, err
	}
	if channels != ChannelsRGB && channels != ChannelsRGBA {
		return 0, 0, fmt.Errorf("bad channels %v: %w", channels, ErrParseHeader)
	}

	var colorSpace uint8
	err = binary.Read(input, binary.BigEndian, &colorSpace)
	if err != nil {
		return 0, 0, err
	}
	if colorSpace != ColorSpaceSRGB && colorSpace != ColorSpaceLinear {
		return 0, 0, fmt.Errorf("bad color space %v: %w", colorSpace, ErrParseHeader)
	}

	return
}

func parseEndMarker(input io.Reader) error {
	var endMarker uint64
	err := binary.Read(input, binary.BigEndian, &endMarker)
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

func parseChunk(input io.Reader, output *image.RGBA) error {
	bs := [4]byte{}
	err := binary.Read(input, binary.BigEndian, &bs)
	if err != nil {
		return err
	}

	output.SetRGBA(0, 0, color.RGBA{bs[1], bs[2], bs[3], 255})
	return nil
}
