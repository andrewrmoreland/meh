package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"strconv"

	_ "golang.org/x/image/webp"
	"golang.org/x/image/draw"
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
