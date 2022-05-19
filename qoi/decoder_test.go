package qoi_test

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"testing"

	"github.com/kropptrevor/go-qoi/qoi"
)

func imageEquals(t *testing.T, expected image.Image, actual image.Image) {
	esize := expected.Bounds().Size()
	asize := actual.Bounds().Size()
	if esize != asize {
		t.Fatalf("expected image size %v but got %v", esize, asize)
	}
	for x := 0; x < esize.X; x++ {
		for y := 0; y < esize.Y; y++ {
			ecol := expected.At(x, y)
			acol := actual.At(x, y)
			if ecol != acol {
				t.Fatalf("expected color %v but got %v at %v", ecol, acol, image.Point{x, y})
			}
		}
	}
}

func TestDecode(t *testing.T) {
	t.Parallel()

	t.Run("Should succeed", func(t *testing.T) {
		t.Parallel()
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, 0, 0, 0, 0, 0, 3, 0,
			0, 0, 0, 0, 0, 0, 0, 1,
		})

		_, err := qoi.Decode(reader)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
	})

	t.Run("Should fail parsing bad magic bytes", func(t *testing.T) {
		t.Parallel()
		reader := bytes.NewReader([]byte{
			'a', 'b', 'c', 'd', 0, 0, 0, 0, 0, 0, 0, 0, 3, 0,
			0, 0, 0, 0, 0, 0, 0, 1,
		})

		_, err := qoi.Decode(reader)

		if err == nil {
			t.Fatal("expected non-nil error")
		}
		expected := qoi.ErrParseHeader
		if !errors.Is(err, expected) {
			t.Fatalf("expected %q but got %q", expected, err)
		}
	})

	t.Run("Should correctly parse header width and height", func(t *testing.T) {
		t.Parallel()
		expectedWidth := 1
		expectedHeight := 1
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, byte(expectedWidth), 0, 0, 0, byte(expectedHeight), 3, 0,
			qoi.TagRGB,
			128, // red
			0,   // green
			0,   // blue
			0, 0, 0, 0, 0, 0, 0, 1,
		})

		m, err := qoi.Decode(reader)

		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}
		actualWidth := m.Bounds().Dx()
		if expectedWidth != actualWidth {
			t.Fatalf("expected %v but got %v", expectedWidth, actualWidth)
		}
		actualHeight := m.Bounds().Dy()
		if expectedHeight != actualHeight {
			t.Fatalf("expected %v but got %v", expectedHeight, actualHeight)
		}
	})

	t.Run("Should fail parsing bad channels", func(t *testing.T) {
		t.Parallel()
		const width = 0
		const height = 0
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, width, 0, 0, 0, height, 9, qoi.ColorSpaceSRGB,
			0, 0, 0, 0, 0, 0, 0, 1,
		})

		_, err := qoi.Decode(reader)

		if err == nil {
			t.Fatal("expected non-nil error")
		}
		expected := qoi.ErrParseHeader
		if !errors.Is(err, expected) {
			t.Fatalf("expected %q but got %q", expected, err)
		}
	})

	t.Run("Should fail parsing bad color space", func(t *testing.T) {
		t.Parallel()
		const width = 0
		const height = 0
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, width, 0, 0, 0, height, qoi.ChannelsRGBA, 2,
			0, 0, 0, 0, 0, 0, 0, 1,
		})

		_, err := qoi.Decode(reader)

		if err == nil {
			t.Fatal("expected non-nil error")
		}
		expected := qoi.ErrParseHeader
		if !errors.Is(err, expected) {
			t.Fatalf("expected %q but got %q", expected, err)
		}
	})

	t.Run("Should fail parsing missing end marker", func(t *testing.T) {
		t.Parallel()
		const width = 0
		const height = 0
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, width, 0, 0, 0, height, qoi.ChannelsRGBA, qoi.ColorSpaceSRGB,
		})

		_, err := qoi.Decode(reader)

		if err == nil {
			t.Fatal("expected non-nil error")
		}
		expected := qoi.ErrParseEndMarker
		if !errors.Is(err, expected) {
			t.Fatalf("expected %q but got %q", expected, err)
		}
	})

	t.Run("Should fail parsing partial end marker", func(t *testing.T) {
		t.Parallel()
		const width = 0
		const height = 0
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, width, 0, 0, 0, height, qoi.ChannelsRGBA, qoi.ColorSpaceSRGB,
			0, 0, 0, 0, 0,
		})

		_, err := qoi.Decode(reader)

		if err == nil {
			t.Fatal("expected non-nil error")
		}
		expected := qoi.ErrParseEndMarker
		if !errors.Is(err, expected) {
			t.Fatalf("expected %q but got %q", expected, err)
		}
	})

	t.Run("Should fail parsing bad end marker", func(t *testing.T) {
		t.Parallel()
		const width = 0
		const height = 0
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, width, 0, 0, 0, height, qoi.ChannelsRGBA, qoi.ColorSpaceSRGB,
			0, 0, 0, 0, 0, 1, 1, 1,
		})

		_, err := qoi.Decode(reader)

		if err == nil {
			t.Fatal("expected non-nil error")
		}
		expected := qoi.ErrParseEndMarker
		if !errors.Is(err, expected) {
			t.Fatalf("expected %q but got %q", expected, err)
		}
	})

	t.Run("Should parse RGB chunk", func(t *testing.T) {
		t.Parallel()
		const size = 1
		expected := image.NewRGBA(image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: size, Y: size},
		})
		expected.SetRGBA(0, 0, color.RGBA{128, 0, 0, 255})
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, size, 0, 0, 0, size, qoi.ChannelsRGBA, qoi.ColorSpaceSRGB,
			qoi.TagRGB,
			128, // red
			0,   // green
			0,   // blue
			0, 0, 0, 0, 0, 0, 0, 1,
		})

		actual, err := qoi.Decode(reader)
		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}

		imageEquals(t, expected, actual)
	})

	t.Run("Should parse RGBA chunk", func(t *testing.T) {
		t.Parallel()
		const size = 1
		expected := image.NewRGBA(image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: size, Y: size},
		})
		expected.SetRGBA(0, 0, color.RGBA{128, 0, 0, 128})
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, size, 0, 0, 0, size, qoi.ChannelsRGBA, qoi.ColorSpaceSRGB,
			qoi.TagRGBA,
			128, // red
			0,   // green
			0,   // blue
			128, // alpha
			0, 0, 0, 0, 0, 0, 0, 1,
		})

		actual, err := qoi.Decode(reader)
		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}

		imageEquals(t, expected, actual)
	})

	t.Run("Should parse index chunk", func(t *testing.T) {
		t.Parallel()
		const width = 3
		const height = 1
		expected := image.NewRGBA(image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: width, Y: height},
		})
		expected.SetRGBA(0, 0, color.RGBA{128, 0, 0, 255})
		expected.SetRGBA(1, 0, color.RGBA{0, 127, 0, 255})
		expected.SetRGBA(2, 0, color.RGBA{128, 0, 0, 255})
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, width, 0, 0, 0, height, qoi.ChannelsRGBA, qoi.ColorSpaceSRGB,
			qoi.TagRGB,
			128, // red
			0,   // green
			0,   // blue
			qoi.TagRGB,
			0,   // red
			127, // green
			0,   // blue
			qoi.TagIndex | 53,
			0, 0, 0, 0, 0, 0, 0, 1,
		})

		actual, err := qoi.Decode(reader)
		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}

		imageEquals(t, expected, actual)
	})

	t.Run("Should parse diff chunk", func(t *testing.T) {
		t.Parallel()
		const width = 2
		const height = 1
		expected := image.NewRGBA(image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: width, Y: height},
		})
		expected.SetRGBA(0, 0, color.RGBA{128, 0, 0, 255})
		expected.SetRGBA(1, 0, color.RGBA{129, 0, 0, 255})
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, width, 0, 0, 0, height, qoi.ChannelsRGBA, qoi.ColorSpaceSRGB,
			qoi.TagRGB,
			128, // red
			0,   // green
			0,   // blue
			qoi.TagDiff | 0b_11_10_10,
			0, 0, 0, 0, 0, 0, 0, 1,
		})

		actual, err := qoi.Decode(reader)
		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}

		imageEquals(t, expected, actual)
	})

	t.Run("Should parse diff chunk with wraparound", func(t *testing.T) {
		t.Parallel()
		const width = 2
		const height = 1
		expected := image.NewRGBA(image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: width, Y: height},
		})
		expected.SetRGBA(0, 0, color.RGBA{128, 255, 0, 255})
		expected.SetRGBA(1, 0, color.RGBA{128, 0, 255, 255})
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, width, 0, 0, 0, height, qoi.ChannelsRGBA, qoi.ColorSpaceSRGB,
			qoi.TagRGB,
			128, // red
			255, // green
			0,   // blue
			qoi.TagDiff | 0b_10_11_01,
			0, 0, 0, 0, 0, 0, 0, 1,
		})

		actual, err := qoi.Decode(reader)
		if err != nil {
			t.Fatalf("expected nil error, but got %v", err)
		}

		imageEquals(t, expected, actual)
	})
}
