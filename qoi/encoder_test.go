package qoi_test

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"io"
	"os"
	"reflect"
	"testing"

	_ "image/png"

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
		if err := binary.Write(expectedBuf, binary.BigEndian, width); err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		if err := binary.Write(expectedBuf, binary.BigEndian, height); err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		if err := binary.Write(expectedBuf, binary.BigEndian, qoi.ChannelsRGBA); err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		if err := binary.Write(expectedBuf, binary.BigEndian, qoi.ColorSpaceSRGB); err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := make([]byte, expectedBuf.Len())
		_, err = buf.Read(actual)
		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		expected, err := io.ReadAll(expectedBuf)
		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

	t.Run("Should have correct end marker", func(t *testing.T) {
		t.Parallel()
		expected := []byte{0, 0, 0, 0, 0, 0, 0, 1}
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[buf.Len()-8:]
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
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
			t.Fatalf("expected %08b, but got %08b", expected, actual)
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
			t.Fatalf("expected %08b, but got %08b", expected, actual)
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
		if expected != actual {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

	t.Run("Should have diff chunk", func(t *testing.T) {
		t.Parallel()
		expected := byte(0b_01_11_10_10)
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		image.SetRGBA(0, 0, color.RGBA{128, 0, 0, 255})
		image.SetRGBA(1, 0, color.RGBA{129, 0, 0, 255})
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[18]
		if expected != actual {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

	t.Run("Should have diff chunk with wraparound", func(t *testing.T) {
		t.Parallel()
		expected := byte(0b_01_10_11_01)
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		image.SetRGBA(0, 0, color.RGBA{128, 255, 0, 255})
		image.SetRGBA(1, 0, color.RGBA{128, 0, 255, 255})
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[18]
		if expected != actual {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

	t.Run("Should have luma chunk", func(t *testing.T) {
		t.Parallel()
		expected := []byte{byte(0b_10_111111), byte(0b_0000_1111)}
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		image.SetRGBA(0, 0, color.RGBA{128, 0, 0, 255})
		image.SetRGBA(1, 0, color.RGBA{151, 31, 38, 255})
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[18:20]
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

	t.Run("Should have luma chunk with wraparound", func(t *testing.T) {
		t.Parallel()
		expected := []byte{byte(0b_10_100010), byte(0b_0110_0101)}
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		image.SetRGBA(0, 0, color.RGBA{128, 255, 0, 255})
		image.SetRGBA(1, 0, color.RGBA{128, 1, 255, 255})
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[18:20]
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

	t.Run("Should have run chunk", func(t *testing.T) {
		t.Parallel()
		expected := byte(0b_11_000010)
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		image.SetRGBA(0, 0, color.RGBA{128, 0, 0, 255})
		image.SetRGBA(1, 0, color.RGBA{128, 0, 0, 255})
		image.SetRGBA(2, 0, color.RGBA{128, 0, 0, 255})
		image.SetRGBA(3, 0, color.RGBA{128, 0, 0, 255})
		image.SetRGBA(4, 0, color.RGBA{128, 129, 0, 255})
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[18]
		if expected != actual {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

	t.Run("Should have max length run chunk", func(t *testing.T) {
		t.Parallel()
		expected := []byte{
			0b11111110, 128, 0, 0, // RGB
			0b_11_111101, // run 62
			0b_11_000000, // run 1
		}
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		for i := 0; i < 64; i++ {
			image.SetRGBA(i, 0, color.RGBA{128, 0, 0, 255})
		}
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[14:20]
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

	t.Run("Should have index chunk after run", func(t *testing.T) {
		t.Parallel()
		expected := []byte{
			0b_11_000001,          // run 2
			0b11111110, 127, 0, 0, // RGB
			0b_00_110101, // index 53
		}
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		image.SetRGBA(0, 0, color.RGBA{0, 0, 0, 255})
		image.SetRGBA(1, 0, color.RGBA{0, 0, 0, 255})
		image.SetRGBA(2, 0, color.RGBA{127, 0, 0, 255})
		image.SetRGBA(3, 0, color.RGBA{0, 0, 0, 255})
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[14:20]
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

	t.Run("Should have run chunk before end marker", func(t *testing.T) {
		t.Parallel()
		expected := byte(0b_11_000000)
		width := uint32(100)
		height := uint32(200)
		image := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		image.SetRGBA(int(width)-2, int(height)-1, color.RGBA{128, 0, 0, 255})
		image.SetRGBA(int(width)-1, int(height)-1, color.RGBA{128, 0, 0, 255})
		var buf bytes.Buffer

		err := qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()[buf.Len()-9]
		if expected != actual {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

	t.Run("Should encode 10x10 correctly", func(t *testing.T) {
		t.Parallel()
		pngFile, err := os.OpenFile("testdata/10x10.png", os.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		image, _, err := image.Decode(pngFile)
		if err != nil {
			t.Fatal(err)
		}
		qoiFile, err := os.OpenFile("testdata/10x10.qoi", os.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := io.ReadAll(qoiFile)
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer

		err = qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

	t.Run("Should encode sample correctly", func(t *testing.T) {
		t.Parallel()
		pngFile, err := os.OpenFile("testdata/sample.png", os.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		image, _, err := image.Decode(pngFile)
		if err != nil {
			t.Fatal(err)
		}
		qoiFile, err := os.OpenFile("testdata/sample.qoi", os.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := io.ReadAll(qoiFile)
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer

		err = qoi.Encode(&buf, image)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actual := buf.Bytes()
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %08b, but got %08b", expected, actual)
		}
	})

}
