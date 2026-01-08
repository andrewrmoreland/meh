package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"golang.org/x/image/draw"
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

// createTestImage creates a test image with gradient colors
func createTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				uint8(x * 255 / width),
				uint8(y * 255 / height),
				128,
				255,
			})
		}
	}
	return img
}

func TestJPEGQuality_AffectsFileSize(t *testing.T) {
	img := createTestImage(100, 100)

	// Encode at low quality
	var lowQualityBuf bytes.Buffer
	jpeg.Encode(&lowQualityBuf, img, &jpeg.Options{Quality: 10})

	// Encode at high quality
	var highQualityBuf bytes.Buffer
	jpeg.Encode(&highQualityBuf, img, &jpeg.Options{Quality: 95})

	lowSize := lowQualityBuf.Len()
	highSize := highQualityBuf.Len()

	if highSize <= lowSize {
		t.Errorf("expected high quality (%d bytes) > low quality (%d bytes)", highSize, lowSize)
	}
}

func TestPNGCompression_AffectsFileSize(t *testing.T) {
	img := createTestImage(100, 100)

	// Encode with no compression
	var noCompressBuf bytes.Buffer
	noCompressEncoder := &png.Encoder{CompressionLevel: png.NoCompression}
	noCompressEncoder.Encode(&noCompressBuf, img)

	// Encode with best compression
	var bestCompressBuf bytes.Buffer
	bestCompressEncoder := &png.Encoder{CompressionLevel: png.BestCompression}
	bestCompressEncoder.Encode(&bestCompressBuf, img)

	noCompressSize := noCompressBuf.Len()
	bestCompressSize := bestCompressBuf.Len()

	if bestCompressSize >= noCompressSize {
		t.Errorf("expected best compression (%d bytes) < no compression (%d bytes)", bestCompressSize, noCompressSize)
	}
}

func TestResize_MaintainsAspectRatio(t *testing.T) {
	// Create 200x100 image (2:1 aspect ratio)
	src := createTestImage(200, 100)

	// Resize to width 100, height should be 50
	dst := image.NewRGBA(image.Rect(0, 0, 100, 50))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	bounds := dst.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 50 {
		t.Errorf("expected 100x50, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestResize_Upscale(t *testing.T) {
	// Create small image
	src := createTestImage(50, 50)

	// Upscale to 200x200
	dst := image.NewRGBA(image.Rect(0, 0, 200, 200))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	bounds := dst.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 200 {
		t.Errorf("expected 200x200, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Verify it's not empty (check a pixel)
	r, g, b, a := dst.At(100, 100).RGBA()
	if a == 0 {
		t.Error("expected non-transparent pixel in upscaled image")
	}
	if r == 0 && g == 0 && b == 0 {
		t.Error("expected non-black pixel in upscaled image")
	}
}

func TestTrimImage_AsymmetricBorder(t *testing.T) {
	// Create image with different border sizes on each side
	img := image.NewRGBA(image.Rect(0, 0, 20, 15))

	// Fill with white
	for y := 0; y < 15; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Content at (5,2) to (15,12) - 10x10 content
	// Left border: 5, Right border: 5, Top border: 2, Bottom border: 3
	green := color.RGBA{0, 255, 0, 255}
	for y := 2; y < 12; y++ {
		for x := 5; x < 15; x++ {
			img.Set(x, y, green)
		}
	}

	result := trimImage(img)
	bounds := result.Bounds()

	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("expected 10x10, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestTrimImage_SinglePixelContent(t *testing.T) {
	// Create image with single non-background pixel
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// Fill with white
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Single red pixel at (5,5)
	img.Set(5, 5, color.RGBA{255, 0, 0, 255})

	result := trimImage(img)
	bounds := result.Bounds()

	if bounds.Dx() != 1 || bounds.Dy() != 1 {
		t.Errorf("expected 1x1, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestMakeBackgroundTransparent_SolidBackground(t *testing.T) {
	// Create 10x10 image with white background and red center
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// Fill with white (background)
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

	result := makeBackgroundTransparent(img)

	// Check that background pixels are now transparent
	_, _, _, a := result.At(0, 0).RGBA()
	if a != 0 {
		t.Errorf("expected transparent pixel at (0,0), got alpha=%d", a)
	}

	// Check that content pixels are preserved
	r, g, b, a := result.At(4, 4).RGBA()
	if a == 0 {
		t.Error("expected opaque pixel at (4,4)")
	}
	if r>>8 != 255 || g != 0 || b != 0 {
		t.Errorf("expected red pixel at (4,4), got r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}

func TestMakeBackgroundTransparent_PreservesNonBackground(t *testing.T) {
	// Create image with multiple colors
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))

	// Fill with blue (background - top-left)
	blue := color.RGBA{0, 0, 255, 255}
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, blue)
		}
	}

	// Add green and red pixels
	img.Set(2, 2, color.RGBA{0, 255, 0, 255}) // green
	img.Set(3, 3, color.RGBA{255, 0, 0, 255}) // red

	result := makeBackgroundTransparent(img)

	// Blue pixels should be transparent
	_, _, _, a := result.At(0, 0).RGBA()
	if a != 0 {
		t.Error("expected blue background to become transparent")
	}

	// Green pixel should remain
	_, g, _, a := result.At(2, 2).RGBA()
	if a == 0 || g>>8 != 255 {
		t.Error("expected green pixel to be preserved")
	}

	// Red pixel should remain
	r, _, _, a := result.At(3, 3).RGBA()
	if a == 0 || r>>8 != 255 {
		t.Error("expected red pixel to be preserved")
	}
}

func TestMakeBackgroundTransparent_AlreadyTransparent(t *testing.T) {
	// Create image with transparent background
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))

	// Fill with transparent
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	// Add opaque pixel
	img.Set(2, 2, color.RGBA{255, 0, 0, 255})

	result := makeBackgroundTransparent(img)

	// Background should remain transparent
	_, _, _, a := result.At(0, 0).RGBA()
	if a != 0 {
		t.Error("expected transparent pixel to remain transparent")
	}

	// Opaque pixel should remain
	_, _, _, a = result.At(2, 2).RGBA()
	if a == 0 {
		t.Error("expected opaque pixel to remain opaque")
	}
}

func TestMakeBackgroundTransparent_AllSameColor(t *testing.T) {
	// Create image that's all one color
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, color.White)
		}
	}

	result := makeBackgroundTransparent(img)

	// All pixels should become transparent
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			_, _, _, a := result.At(x, y).RGBA()
			if a != 0 {
				t.Errorf("expected all pixels transparent, but (%d,%d) has alpha=%d", x, y, a)
			}
		}
	}
}

func TestMakeBackgroundTransparent_PreservesDimensions(t *testing.T) {
	img := createTestImage(100, 50)

	result := makeBackgroundTransparent(img)
	bounds := result.Bounds()

	if bounds.Dx() != 100 || bounds.Dy() != 50 {
		t.Errorf("expected 100x50, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}
