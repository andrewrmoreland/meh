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
├── cmd/
│   └── main.go               # WASM entry point
├── imaging/
│   ├── imaging.go            # Image processing functions
│   └── imaging_test.go       # Tests
├── web/
│   ├── index.html            # Web interface
│   ├── main.wasm             # Built WASM binary (generated)
│   └── wasm_exec.js          # Go WASM runtime (generated)
├── build-wasm.sh             # Build script
├── .github/workflows/ci.yml  # CI/CD
├── go.mod                    # Module definition
└── go.sum                    # Dependency checksums
```

## Quick Commands

```bash
# Run tests
go test -v ./...

# Build WASM
./build-wasm.sh

# Serve locally
cd web && python3 -m http.server 8080
```

## Key Source Files

### `imaging/imaging.go` - Image Processing

Exported functions:
- **`Trim(img)`** - Removes borders (transparent or solid color)
- **`RemoveBackground(img)`** - Flood-fill background removal

### `cmd/main.go` - WASM Entry Point

JavaScript-callable via `processImage()` function.

**processImage() Parameters:**
1. `args[0]`: Uint8Array image data
2. `args[1]`: target width (int)
3. `args[2]`: target height (int)
4. `args[3]`: trim flag (bool)
5. `args[4]`: format string ("png" or "jpeg")
6. `args[5]`: quality int (1-100)
7. `args[6]`: transparentBg flag (bool, optional)

## Testing

```bash
go test -v ./...
```

Test file: `imaging/imaging_test.go`

## Dependencies

```
golang.org/x/image v0.34.0
```

## CI/CD

1. **Test** - Runs `go test -v ./...`
2. **Build WASM** - Compiles to `web/main.wasm`
3. **Deploy** - GitHub Pages (main branch only)
