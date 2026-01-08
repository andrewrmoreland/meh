package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/resize", resizeHandler)

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Image Resizer</title>
</head>
<body>
    <h1>Image Resizer</h1>
    <form action="/resize" method="post" enctype="multipart/form-data">
        <p>
            <label>Image: <input type="file" name="image" accept="image/*" required></label>
        </p>
        <p>
            <label>Width: <input type="number" name="width" placeholder="e.g. 200"></label>
        </p>
        <p>
            <label>Height: <input type="number" name="height" placeholder="e.g. 200"></label>
        </p>
        <p>
            <label>Output format:
                <select name="format">
                    <option value="png">PNG</option>
                    <option value="jpeg">JPEG</option>
                </select>
            </label>
        </p>
        <p>
            <label><input type="checkbox" name="trim" value="1"> Trim borders (transparent or solid color)</label>
        </p>
        <p>
            <label><input type="checkbox" name="transparentBg" value="1"> Make background transparent (uses top-left pixel color, PNG only)</label>
        </p>
        <p>
            <button type="submit">Resize</button>
        </p>
    </form>
    <p><small>Supports PNG, JPEG, and WebP input. Leave width or height empty to maintain aspect ratio.</small></p>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func resizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode image: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Apply trim if requested
	if r.FormValue("trim") == "1" {
		img = trimImage(img)
	}

	// Make background transparent if requested
	if r.FormValue("transparentBg") == "1" {
		img = makeBackgroundTransparent(img)
	}

	// Get dimensions
	widthStr := r.FormValue("width")
	heightStr := r.FormValue("height")
	format := r.FormValue("format")

	if format == "" {
		format = "png"
	}

	// Calculate new dimensions
	origBounds := img.Bounds()
	origWidth := origBounds.Dx()
	origHeight := origBounds.Dy()

	newWidth, _ := strconv.Atoi(widthStr)
	newHeight, _ := strconv.Atoi(heightStr)

	// Maintain aspect ratio if only one dimension is provided
	if newWidth > 0 && newHeight == 0 {
		newHeight = int(float64(origHeight) * float64(newWidth) / float64(origWidth))
	} else if newHeight > 0 && newWidth == 0 {
		newWidth = int(float64(origWidth) * float64(newHeight) / float64(origHeight))
	} else if newWidth == 0 && newHeight == 0 {
		// Default to original size if no dimensions provided
		newWidth = origWidth
		newHeight = origHeight
	}

	// Resize the image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	// Encode and send the response
	switch format {
	case "jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Disposition", "attachment; filename=resized.jpg")
		jpeg.Encode(w, dst, &jpeg.Options{Quality: 90})
	default:
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", "attachment; filename=resized.png")
		png.Encode(w, dst)
	}
}

// trimImage removes transparent borders (if image has transparency) or solid color borders
// (using top-left pixel as reference). Returns the cropped subimage.
func trimImage(img image.Image) image.Image {
	bounds := img.Bounds()
	minX, minY := bounds.Min.X, bounds.Min.Y
	maxX, maxY := bounds.Max.X, bounds.Max.Y

	// Check if image has transparency by sampling top-left pixel
	topLeft := img.At(minX, minY)
	_, _, _, a := topLeft.RGBA()
	hasTransparency := a < 0xffff

	// Determine if a pixel should be trimmed
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

// colorsEqual compares two colors for exact equality
func colorsEqual(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

// makeBackgroundTransparent replaces background pixels with transparent pixels.
// Only pixels connected to the image edges are considered background (flood-fill from borders).
func makeBackgroundTransparent(img image.Image) image.Image {
	bounds := img.Bounds()
	bgColor := img.At(bounds.Min.X, bounds.Min.Y)
	width := bounds.Dx()
	height := bounds.Dy()

	// Track which pixels are background (connected to edges)
	isBackground := make([][]bool, height)
	for i := range isBackground {
		isBackground[i] = make([]bool, width)
	}

	// Flood-fill from all edge pixels that match the background color
	type point struct{ x, y int }
	queue := make([]point, 0)

	// Add all edge pixels matching background color to the queue
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		// Top edge
		if colorsEqual(img.At(x, bounds.Min.Y), bgColor) {
			queue = append(queue, point{x - bounds.Min.X, 0})
			isBackground[0][x-bounds.Min.X] = true
		}
		// Bottom edge
		if colorsEqual(img.At(x, bounds.Max.Y-1), bgColor) {
			queue = append(queue, point{x - bounds.Min.X, height - 1})
			isBackground[height-1][x-bounds.Min.X] = true
		}
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		// Left edge
		if colorsEqual(img.At(bounds.Min.X, y), bgColor) {
			queue = append(queue, point{0, y - bounds.Min.Y})
			isBackground[y-bounds.Min.Y][0] = true
		}
		// Right edge
		if colorsEqual(img.At(bounds.Max.X-1, y), bgColor) {
			queue = append(queue, point{width - 1, y - bounds.Min.Y})
			isBackground[y-bounds.Min.Y][width-1] = true
		}
	}

	// BFS flood-fill
	dirs := []point{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]

		for _, d := range dirs {
			nx, ny := p.x+d.x, p.y+d.y
			if nx >= 0 && nx < width && ny >= 0 && ny < height && !isBackground[ny][nx] {
				if colorsEqual(img.At(nx+bounds.Min.X, ny+bounds.Min.Y), bgColor) {
					isBackground[ny][nx] = true
					queue = append(queue, point{nx, ny})
				}
			}
		}
	}

	// Create result image
	result := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if isBackground[y-bounds.Min.Y][x-bounds.Min.X] {
				result.Set(x, y, color.Transparent)
			} else {
				result.Set(x, y, img.At(x, y))
			}
		}
	}

	return result
}
