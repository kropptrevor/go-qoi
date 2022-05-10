package qoi_test

import (
	"bytes"
	"encoding/binary"
	"image"
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
}
