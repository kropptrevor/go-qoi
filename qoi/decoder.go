package qoi

import (
	"encoding/binary"
	"errors"
	"image"
	"io"
)

var ErrParseHeader = errors.New("failed to parse QOI header")

func Decode(reader io.Reader) (image.Image, error) {
	magic := make([]byte, 4)
	err := binary.Read(reader, binary.BigEndian, magic)
	if err != nil {
		return image.NewNRGBA(image.Rectangle{}), err
	}
	correctMagic := []byte{'q', 'o', 'i', 'f'}
	if string(magic) != string(correctMagic) {
		return image.NewNRGBA(image.Rectangle{}), ErrParseHeader
	}
	var width uint32
	err = binary.Read(reader, binary.BigEndian, &width)
	if err != nil {
		return image.NewNRGBA(image.Rectangle{}), err
	}
	var height uint32
	err = binary.Read(reader, binary.BigEndian, &height)
	if err != nil {
		return image.NewNRGBA(image.Rectangle{}), err
	}
	return image.NewNRGBA(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(width), int(height)},
	}), nil
}
