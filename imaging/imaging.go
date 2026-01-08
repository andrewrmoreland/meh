// Package imaging provides image processing functions for resizing,
// trimming borders, and background transparency manipulation.
package imaging

import (
	"image"
	"image/color"

	"golang.org/x/image/draw"
)

// Trim removes transparent borders (if image has transparency) or solid color borders.
func Trim(img image.Image) image.Image {
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

// colorsEqual compares two colors for equality.
func colorsEqual(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

// RemoveBackground replaces background pixels with transparent pixels.
// Only pixels connected to the image edges are considered background (flood-fill from borders).
func RemoveBackground(img image.Image) image.Image {
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
