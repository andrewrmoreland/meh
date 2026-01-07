package main

import (
	"image"
	"image/color"
	"testing"
)

func TestColorsEqual(t *testing.T) {
	tests := []struct {
		name string
		c1   color.Color
		c2   color.Color
		want bool
	}{
		{"same white", color.White, color.White, true},
		{"same black", color.Black, color.Black, true},
		{"white vs black", color.White, color.Black, false},
		{"same RGBA", color.RGBA{255, 128, 64, 255}, color.RGBA{255, 128, 64, 255}, true},
		{"different RGBA", color.RGBA{255, 128, 64, 255}, color.RGBA{255, 128, 65, 255}, false},
		{"same transparent", color.RGBA{0, 0, 0, 0}, color.RGBA{0, 0, 0, 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorsEqual(tt.c1, tt.c2)
			if got != tt.want {
				t.Errorf("colorsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimImage_SolidBorder(t *testing.T) {
	// Create 10x10 image with white border and red center (4x4)
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// Fill with white
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Red center at (3,3) to (6,6)
	red := color.RGBA{255, 0, 0, 255}
	for y := 3; y < 7; y++ {
		for x := 3; x < 7; x++ {
			img.Set(x, y, red)
		}
	}

	result := trimImage(img)
	bounds := result.Bounds()

	if bounds.Dx() != 4 || bounds.Dy() != 4 {
		t.Errorf("expected 4x4, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Check that the result contains red pixels
	r, g, b, _ := result.At(0, 0).RGBA()
	if r>>8 != 255 || g != 0 || b != 0 {
		t.Errorf("expected red pixel at (0,0), got r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}

func TestTrimImage_TransparentBorder(t *testing.T) {
	// Create 10x10 image with transparent border and opaque center
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// Fill with transparent
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	// Blue center at (2,2) to (7,7)
	blue := color.RGBA{0, 0, 255, 255}
	for y := 2; y < 8; y++ {
		for x := 2; x < 8; x++ {
			img.Set(x, y, blue)
		}
	}

	result := trimImage(img)
	bounds := result.Bounds()

	if bounds.Dx() != 6 || bounds.Dy() != 6 {
		t.Errorf("expected 6x6, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestTrimImage_NoTrimNeeded(t *testing.T) {
	// Create image with no uniform border
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))

	// Fill with different colors
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 50), uint8(y * 50), 128, 255})
		}
	}

	result := trimImage(img)
	bounds := result.Bounds()

	// Should remain 5x5 since top-left pixel doesn't match others
	if bounds.Dx() != 5 || bounds.Dy() != 5 {
		t.Errorf("expected 5x5 (no trim), got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestTrimImage_AllSameColor(t *testing.T) {
	// Create image that's all one color
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, color.White)
		}
	}

	result := trimImage(img)
	bounds := result.Bounds()

	// Edge case: all same color means everything could be trimmed
	// Current implementation returns original if nothing found
	if bounds.Dx() != 5 || bounds.Dy() != 5 {
		t.Errorf("expected 5x5 (no content to keep), got %dx%d", bounds.Dx(), bounds.Dy())
	}
}
