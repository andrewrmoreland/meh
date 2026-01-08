# CLAUDE.md - AI Assistant Guide for Image Resizer

This document provides guidance for AI assistants working with this codebase.

## Project Overview

**Project:** Image Resizer (WASM)
**Language:** Go 1.24
**Type:** Browser-based image processing via WebAssembly
**Purpose:** Provides image resizing, trimming, and background transparency manipulation in the browser

## Directory Structure

```
/
├── imgproc/
│   ├── imgproc.go            # Shared image processing functions
│   └── imgproc_test.go       # Tests for image processing
├── wasm/
│   ├── main.go               # WASM entry point
│   ├── index.html            # Web interface
│   └── wasm_exec.js          # Go WASM runtime (generated)
├── build-wasm.sh             # Build script for WASM
├── .github/workflows/ci.yml  # GitHub Actions CI/CD
├── go.mod                    # Module definition
├── go.sum                    # Dependency checksums
└── .gitignore                # Ignores: wasm/main.wasm
```

## Quick Commands

```bash
# Run tests
go test -v ./...

# Build WASM
./build-wasm.sh

# Serve WASM locally
cd wasm && python3 -m http.server 8080
```

## Key Source Files

### `imgproc/imgproc.go` - Shared Image Processing

Contains the core image processing functions:

- **`TrimImage(img)`** - Removes borders (transparent or solid color)
- **`MakeBackgroundTransparent(img)`** - Flood-fill algorithm for background removal
- **`ColorsEqual(c1, c2)`** - Color comparison utility

### `wasm/main.go` - WASM Entry Point

Browser-based implementation with:
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

**Test file:** `imgproc/imgproc_test.go`

**Test organization:**
- `TestColorsEqual` - Color comparison (6 test cases)
- `TestTrimImage_*` - Border trimming (7 test functions)
- `TestMakeBackgroundTransparent_*` - Transparency function (6 test functions)

**Testing patterns:**
- Table-driven tests for multiple cases
- `createTestImage()` for fixture generation
- Pixel verification via direct sampling

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
- **Exported Functions:** PascalCase (`TrimImage`, `MakeBackgroundTransparent`)
- **Internal variables:** camelCase with descriptive names

### Error Handling
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
- Single-threaded processing

## Development Notes

1. **Run tests before committing** - `go test -v ./...`

2. **Test WASM locally** - Build with `./build-wasm.sh` and serve locally to verify browser behavior.

3. **Add tests for new features** - Follow existing patterns in `imgproc/imgproc_test.go`.

4. **Edge cases to consider:**
   - Single-pixel images
   - All-same-color images
   - Asymmetric borders
   - Images with existing transparency
   - Both PNG and JPEG output formats
