package qoi_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/kropptrevor/go-qoi/qoi"
)

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
		expectedWidth := 3
		expectedHeight := 5
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, byte(expectedWidth), 0, 0, 0, byte(expectedHeight), 3, 0,
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
		const width = 2
		const height = 2
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
		const width = 2
		const height = 2
		reader := bytes.NewReader([]byte{
			'q', 'o', 'i', 'f', 0, 0, 0, width, 0, 0, 0, height, qoi.ChannelRGBA, 2,
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
}
