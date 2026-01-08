# CLAUDE.md - AI Assistant Guide for Image Resizer

This document provides comprehensive guidance for AI assistants working with this codebase.

## Project Overview

**Project:** Image Resizer
**Language:** Go 1.25
**Type:** Image processing application with dual deployment modes
**Purpose:** Provides image resizing, trimming, and background transparency manipulation capabilities

The project offers two implementations:
- **Server-side**: HTTP web server for image processing (port 8080)
- **Client-side**: WebAssembly (WASM) for browser-based processing

## Directory Structure

```
/
├── main.go                    # HTTP server implementation
├── main_test.go               # Server-side tests
├── wasm/
│   ├── main.go               # WASM implementation
│   ├── index.html            # WASM web interface
│   └── wasm_exec.js          # Go WASM runtime (generated)
├── build-wasm.sh             # Build script for WASM
├── .github/workflows/ci.yml   # GitHub Actions CI/CD
├── go.mod                     # Module definition
├── go.sum                     # Dependency checksums
└── .gitignore                # Ignores: image-resizer binary, wasm/main.wasm
```

## Quick Commands

```bash
# Run tests
go test -v ./...

# Build server binary
go build

# Run server (after building)
./image-resizer

# Build WASM
./build-wasm.sh

# Serve WASM locally
cd wasm && python3 -m http.server 8080
```

## Key Source Files

### `main.go` - HTTP Server

Main server implementation with:
- **Routes:**
  - `GET /` - Serves HTML form UI
  - `POST /resize` - Processes uploaded images

- **Form Parameters:**
  - `image` (file) - Input image (PNG, JPEG, WebP)
  - `width` (int) - Target width (optional)
  - `height` (int) - Target height (optional)
  - `format` (string) - Output format: "png" or "jpeg"
  - `trim` (checkbox) - Enable border trimming
  - `transparentBg` (checkbox) - Make background transparent

- **Key Functions:**
  - `resizeHandler()` - Main HTTP request handler
  - `trimImage()` - Removes borders (transparent or solid color)
  - `makeBackgroundTransparent()` - Flood-fill algorithm for background removal
  - `colorsEqual()` - Color comparison utility

### `wasm/main.go` - WASM Implementation

Browser-based implementation with:
- Same image processing logic as server
- JavaScript-callable via `processImage()` function
- Build tag: `//go:build js && wasm`

**processImage() Parameters:**
1. `args[0]`: Uint8Array image data
2. `args[1]`: target width (int)
3. `args[2]`: target height (int)
4. `args[3]`: trim flag (bool)
5. `args[4]`: format string ("png" or "jpeg")
6. `args[5]`: quality int (1-100)
7. `args[6]`: transparentBg flag (bool, optional)

## Core Algorithms

### Aspect Ratio Calculation

```
If width only:  new_height = old_height * (new_width / old_width)
If height only: new_width = old_width * (new_height / old_height)
If neither:     Use original dimensions
If both:        Use provided dimensions
```

### Image Scaling

Uses CatmullRom interpolation via `golang.org/x/image/draw` for high-quality resampling.

### Trim Image Algorithm

1. Detects if image has transparency by sampling top-left pixel
2. Determines trim mode: transparent pixels (alpha==0) or matching top-left color
3. Scans all four edges inward to find content boundaries
4. Returns cropped subimage or original if no trim needed

### Background Transparency Algorithm

1. Samples top-left pixel as background color reference
2. BFS flood-fill from all edge pixels matching background color
3. Only marks pixels as background if connected to image edges
4. Interior pixels matching background color are preserved

## Testing

**Test file:** `main_test.go`

**Test organization:**
- `TestColorsEqual` - Color comparison (6 test cases)
- `TestTrimImage_*` - Border trimming (5 test functions)
- `TestMakeBackgroundTransparent_*` - Transparency function (6 test functions)
- `TestJPEGQuality_AffectsFileSize` - Compression validation
- `TestPNGCompression_AffectsFileSize` - PNG encoding levels
- `TestResize_*` - Resizing and aspect ratio (2 test functions)

**Testing patterns:**
- Table-driven tests for multiple cases
- `createTestImage()` for fixture generation
- Pixel verification via direct sampling
- File size assertions for compression tests

## Dependencies

Single external dependency:
```
golang.org/x/image v0.34.0
```

Used for:
- Image decoding (PNG, JPEG, WebP)
- CatmullRom scaling algorithm
- WebP format support

## CI/CD Pipeline

GitHub Actions workflow (`.github/workflows/ci.yml`):

1. **Test Job** - Runs `go test -v ./...`
2. **Build WASM Job** - Compiles WASM binary
3. **Deploy Job** - Deploys to GitHub Pages (main branch only)

## Code Conventions

### Naming
- **Functions:** CamelCase (`trimImage`, `makeBackgroundTransparent`)
- **Variables:** camelCase with descriptive names (`shouldTrim`, `newWidth`)

### Error Handling
- HTTP: Descriptive error messages returned to client
- WASM: Errors wrapped in return map: `{"error": "message"}`
- Early returns with descriptive error messages

### Image Processing Patterns
- Use `image.Bounds()` for dimension handling
- Create output with `image.NewRGBA()`
- Use `draw.Scale()` and `draw.Copy()` for pixel manipulation
- Use `color.Color` interface for abstraction

### Testing Patterns
- Use `t.Run()` for sub-tests
- Create fixtures with helper functions
- Direct value comparisons with descriptive error messages

## Known Limitations

- Transparent background uses top-left pixel as reference color only
- Maximum upload size: 10MB (hard-coded)
- Single-threaded processing
- JPEG output quality fixed at 90% (server-side)

## Development Notes

When making changes:

1. **Keep implementations in sync** - Both `main.go` and `wasm/main.go` share image processing logic. Changes to algorithms should be reflected in both.

2. **Run tests before committing** - `go test -v ./...`

3. **Test WASM locally** - Build with `./build-wasm.sh` and serve locally to verify browser behavior.

4. **Add tests for new features** - Follow existing patterns in `main_test.go`.

5. **Edge cases to consider:**
   - Single-pixel images
   - All-same-color images
   - Asymmetric borders
   - Images with existing transparency
   - Both PNG and JPEG output formats
