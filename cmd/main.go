//go:build js && wasm

package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"syscall/js"

	"image-resizer/imaging"

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
// Args: imageData (Uint8Array), width (int), height (int), trim (bool), format (string), quality (int), transparentBg (bool)
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
	transparentBg := false
	if len(args) >= 7 {
		transparentBg = args[6].Bool()
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return map[string]interface{}{"error": "failed to decode image: " + err.Error()}
	}

	// Apply trim if requested
	if trim {
		img = imaging.Trim(img)
	}

	// Make background transparent if requested
	if transparentBg {
		img = imaging.RemoveBackground(img)
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
		// Map quality to PNG compression level (inverted: 1-25 = BestCompression, 76-100 = NoCompression)
		// This makes "higher = faster/larger" consistent with JPEG's "higher = better/larger"
		var compression png.CompressionLevel
		if quality <= 25 {
			compression = png.BestCompression
		} else if quality <= 50 {
			compression = png.DefaultCompression
		} else if quality <= 75 {
			compression = png.BestSpeed
		} else {
			compression = png.NoCompression
		}
		encoder := &png.Encoder{CompressionLevel: compression}
		err = encoder.Encode(&buf, dst)
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
