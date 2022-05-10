package qoi_test

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"io"
	"reflect"
	"testing"

	"github.com/kropptrevor/go-qoi/qoi"
)

func TestEncode(t *testing.T) {
	t.Parallel()

	t.Run("Should succeed", func(t *testing.T) {
		t.Parallel()
		image := image.NewRGBA(image.Rect(0, 0, 100, 200))
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
	})

	t.Run("Should have correct header", func(t *testing.T) {
		t.Parallel()
		expectedBuf := bytes.NewBuffer([]byte{'q', 'o', 'i', 'f'})
		width := uint32(100)
		height := uint32(200)
		binary.Write(expectedBuf, binary.BigEndian, width)
		binary.Write(expectedBuf, binary.BigEndian, height)
		binary.Write(expectedBuf, binary.BigEndian, qoi.ChannelRGBA)
		binary.Write(expectedBuf, binary.BigEndian, qoi.ColorSpaceSRGB)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		var buf bytes.Buffer

		qoi.Encode(&buf, image)

		readBuf := make([]byte, expectedBuf.Len())
		_, err := buf.Read(readBuf)
		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		expected, err := io.ReadAll(expectedBuf)
		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		if !reflect.DeepEqual(expected, readBuf) {
			t.Fatalf("expected %v, but got %v", expected, readBuf)
		}
	})

	t.Run("Should have correct end marker", func(t *testing.T) {
		t.Parallel()
		expected := []byte{0, 0, 0, 0, 0, 0, 0, 1}
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		var buf bytes.Buffer

		qoi.Encode(&buf, image)

		actual := buf.Bytes()[buf.Len()-8:]
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, but got %v", expected, actual)
		}
	})

	t.Run("Should have RGBA chunk", func(t *testing.T) {
		t.Parallel()
		expected := []byte{0b11111111, 0, 0, 0, 128}
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		image.SetRGBA(0, 0, color.RGBA{0, 0, 0, 128})
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[14:19]
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, but got %v", expected, actual)
		}
	})

	t.Run("Should have RGB chunk", func(t *testing.T) {
		t.Parallel()
		expected := []byte{0b11111110, 128, 0, 0}
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		image.SetRGBA(0, 0, color.RGBA{128, 0, 0, 255})
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[14:18]
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, but got %v", expected, actual)
		}
	})

	t.Run("Should have index chunk", func(t *testing.T) {
		t.Parallel()
		expected := byte(53)
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		image.SetRGBA(0, 0, color.RGBA{128, 0, 0, 255})
		image.SetRGBA(1, 0, color.RGBA{0, 127, 0, 255})
		image.SetRGBA(2, 0, color.RGBA{128, 0, 0, 255})
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[22]
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, but got %v", expected, actual)
		}
	})
}
