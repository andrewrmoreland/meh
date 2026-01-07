//go:build js && wasm

package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"syscall/js"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

func main() {
	// Register functions for JavaScript to call
	js.Global().Set("processImage", js.FuncOf(processImage))

	// Keep the program running
	select {}
}

// processImage is called from JavaScript with image data and options
// Args: imageData (Uint8Array), width (int), height (int), trim (bool), format (string), quality (int)
// Returns: processed image as Uint8Array
func processImage(this js.Value, args []js.Value) interface{} {
	if len(args) < 6 {
		return map[string]interface{}{"error": "missing arguments"}
	}

	// Get image data from JavaScript Uint8Array
	jsData := args[0]
	length := jsData.Get("length").Int()
	imageData := make([]byte, length)
	js.CopyBytesToGo(imageData, jsData)

	width := args[1].Int()
	height := args[2].Int()
	trim := args[3].Bool()
	format := args[4].String()
	quality := args[5].Int()
	if quality <= 0 || quality > 100 {
		quality = 90
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return map[string]interface{}{"error": "failed to decode image: " + err.Error()}
	}

	// Apply trim if requested
	if trim {
		img = trimImage(img)
	}

	// Calculate new dimensions
	origBounds := img.Bounds()
	origWidth := origBounds.Dx()
	origHeight := origBounds.Dy()

	newWidth := width
	newHeight := height

	// Maintain aspect ratio if only one dimension is provided
	if newWidth > 0 && newHeight == 0 {
		newHeight = int(float64(origHeight) * float64(newWidth) / float64(origWidth))
	} else if newHeight > 0 && newWidth == 0 {
		newWidth = int(float64(origWidth) * float64(newHeight) / float64(origHeight))
	} else if newWidth == 0 && newHeight == 0 {
		newWidth = origWidth
		newHeight = origHeight
	}

	// Resize the image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	// Encode the result
	var buf bytes.Buffer
	var mimeType string

	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: quality})
		mimeType = "image/jpeg"
	default:
		err = png.Encode(&buf, dst)
		mimeType = "image/png"
	}

	if err != nil {
		return map[string]interface{}{"error": "failed to encode image: " + err.Error()}
	}

	// Create Uint8Array to return to JavaScript
	result := buf.Bytes()
	jsResult := js.Global().Get("Uint8Array").New(len(result))
	js.CopyBytesToJS(jsResult, result)

	return map[string]interface{}{
		"data":     jsResult,
		"mimeType": mimeType,
		"width":    newWidth,
		"height":   newHeight,
		"size":     len(result),
	}
}

// trimImage removes transparent borders (if image has transparency) or solid color borders
func trimImage(img image.Image) image.Image {
	bounds := img.Bounds()
	minX, minY := bounds.Min.X, bounds.Min.Y
	maxX, maxY := bounds.Max.X, bounds.Max.Y

	// Check if image has transparency by sampling top-left pixel
	topLeft := img.At(minX, minY)
	_, _, _, a := topLeft.RGBA()
	hasTransparency := a < 0xffff

	shouldTrim := func(x, y int) bool {
		c := img.At(x, y)
		if hasTransparency {
			_, _, _, alpha := c.RGBA()
			return alpha == 0
		}
		return colorsEqual(c, topLeft)
	}

	// Find top edge
	top := minY
	for y := minY; y < maxY; y++ {
		found := false
		for x := minX; x < maxX; x++ {
			if !shouldTrim(x, y) {
				found = true
				break
			}
		}
		if found {
			top = y
			break
		}
	}

	// Find bottom edge
	bottom := maxY
	for y := maxY - 1; y >= top; y-- {
		found := false
		for x := minX; x < maxX; x++ {
			if !shouldTrim(x, y) {
				found = true
				break
			}
		}
		if found {
			bottom = y + 1
			break
		}
	}

	// Find left edge
	left := minX
	for x := minX; x < maxX; x++ {
		found := false
		for y := top; y < bottom; y++ {
			if !shouldTrim(x, y) {
				found = true
				break
			}
		}
		if found {
			left = x
			break
		}
	}

	// Find right edge
	right := maxX
	for x := maxX - 1; x >= left; x-- {
		found := false
		for y := top; y < bottom; y++ {
			if !shouldTrim(x, y) {
				found = true
				break
			}
		}
		if found {
			right = x + 1
			break
		}
	}

	// If nothing to trim, return original
	if left == minX && right == maxX && top == minY && bottom == maxY {
		return img
	}

	// Create cropped image
	cropped := image.NewRGBA(image.Rect(0, 0, right-left, bottom-top))
	draw.Copy(cropped, image.Point{}, img, image.Rect(left, top, right, bottom), draw.Src, nil)
	return cropped
}

func colorsEqual(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
