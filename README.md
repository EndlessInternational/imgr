# imgr

A minimal, fast image manipulation tool designed for programmatic use and LLM tool integration.

## Overview

imgr is a focused image processing tool built for reliability, speed, and minimal resource usage. It excels in automated workflows, scripts, and as a tool for AI assistants where low memory footprint and consistent behavior matter.

**Key characteristics:**
- **Fast** - Single binary with minimal startup time.
- **Reliable** - Defensive error handling with clear error messages.
- **Efficient** - Low memory footprint enables concurrent use by multiple LLM instances.
- **Predictable** - Consistent behavior across platforms.
- **Focused** - Core image operations without unnecessary features.

## Installation

### Prerequisites

```bash
# macOS
brew install libheif

# Ubuntu/Debian
sudo apt-get install libheif-dev

# Fedora
sudo dnf install libheif-devel
```

**Note:** libheif is only required for HEIC/HEIF support. All other formats work without any system dependencies.

### Build from source

```bash
git clone https://github.com/yourusername/imgr.git
cd imgr
./build.sh
```

This creates `build/imgr` - a single binary you can copy anywhere.

### Install

```bash
cp build/imgr /usr/local/bin/
```

## Design Goals

### Built for LLM Tool Integration

imgr is designed to be called by AI assistants and automated systems:

- **Low dependencies** - Minimal system requirements reduce deployment complexity.
- **Fast execution** - Quick startup and processing enable responsive tool calls.
- **Small memory footprint** - Multiple LLM instances can use the tool concurrently without resource contention.
- **Clear error messages** - Full English sentences improve LLM understanding and error recovery.
- **Predictable behavior** - Consistent results across different inputs enable reliable automation.
- **Simple CLI** - Straightforward command structure makes it easy to generate correct invocations.
- **JSON output** - Optional `--json` flag enables structured output for programmatic use.

### Core Features

**Supported Formats:**  
Read: JPEG, PNG, GIF, TIFF, WebP, HEIF/HEIC  
Write: JPEG, PNG, GIF, TIFF

**Smart Resizing:**
- The tool maintains aspect ratio by default.
- Images fit within specified bounds when both dimensions are provided.
- The optional `--no-enlarge` flag prevents upscaling of small images.
- High-quality bilinear interpolation ensures good visual results.

## Usage

### Quick Start

```bash
# Convert JPEG to PNG
imgr transform photo.jpg photo.png

# Resize to 800px wide (maintains aspect ratio)
imgr transform -w 800 large.jpg small.jpg

# Fit within 1920x1080 (maintains aspect ratio)
imgr transform -w 1920 -h 1080 photo.jpg wallpaper.jpg

# Get image information
imgr info photo.jpg

# Get image information as JSON
imgr --json info photo.jpg
```

### Global Flags

- `--json` - Output results as JSON for programmatic use.
- `--help` - Show help information.
- `--version` - Show version information.

### Commands

#### transform

Resize or convert images.

```bash
imgr transform [options] <input> <output>
```

**Flags:**
- `-w, --width N` - Sets the output width in pixels (or maximum width when height is also specified).
- `-h, --height N` - Sets the output height in pixels (or maximum height when width is also specified).
- `-q, --quality N` - Sets the JPEG quality from 0 to 100 (default: 90).
- `--no-enlarge` - Prevents the output image from being larger than the source image.

**Examples:**

```bash
# Width only - scales to that width
imgr transform -w 800 photo.jpg thumbnail.jpg

# Height only - scales to that height
imgr transform -h 600 photo.jpg thumbnail.jpg

# Both dimensions - fits within bounds (maintains aspect)
imgr transform -w 1920 -h 1080 photo.jpg wallpaper.jpg
# 2000x1500 → 1440x1080 (fits within box, maintains aspect)
# 1600x1200 → 1440x1080 (fits width)

# Prevent upscaling
imgr transform -w 2000 -h 2000 --no-enlarge small.jpg output.jpg
# 800x600 stays 800x600 (no enlargement)

# High quality JPEG
imgr transform -w 1920 -q 95 photo.jpg high-quality.jpg

# Format conversion
imgr transform photo.heic photo.jpg          # HEIC → JPEG
imgr transform screenshot.png graphic.jpg    # PNG → JPEG
imgr transform photo.jpg lossless.png        # JPEG → PNG
```

#### info

Display detailed information about an image.

```bash
imgr info <input>
```

**Output:**
```
File:         photo.jpg
Path:         /Users/you/photos/photo.jpg
Format:       JPEG
Dimensions:   1920 × 1080 pixels
Aspect Ratio: 1.78:1
Transparency: false
Color Model:  YCbCr
File Size:    245680 bytes (239.92 KB)
```

### JSON Output

Use the `--json` flag for structured output, useful when calling imgr from scripts or other programs.

**Success response:**
```json
{
  "success": true,
  "data": {
    "file": "photo.jpg",
    "path": "/path/to/photo.jpg",
    "format": "jpeg",
    "width": 1920,
    "height": 1080,
    "aspect_ratio": 1.78,
    "has_alpha": false,
    "color_model": "YCbCr",
    "file_size_bytes": 245680,
    "file_size_kb": 239.92
  }
}
```

**Error response:**
```json
{
  "success": false,
  "error": {
    "message": "The file photo.jpg could not be accessed: open photo.jpg: no such file or directory."
  }
}
```

### Common Use Cases

#### Convert iPhone HEIC photos

```bash
# Single file
imgr transform photo.heic photo.jpg

# Batch convert all HEIC files
for file in *.heic; do
  imgr transform "$file" "${file%.heic}.jpg"
done

# Resize while converting
for file in *.heic; do
  imgr transform -w 1920 "$file" "${file%.heic}.jpg"
done
```

#### Create thumbnails

```bash
# Single thumbnail
imgr transform -w 400 -h 400 photo.jpg thumb.jpg

# Batch create thumbnails
for file in *.jpg; do
  imgr transform -w 300 -h 300 "$file" "thumb_$file"
done
```

#### Optimize for web

```bash
# Reduce size with quality setting
imgr transform -w 1200 -q 85 large.jpg web.jpg

# Fit within common screen sizes
imgr transform -w 1920 -h 1080 --no-enlarge photo.jpg optimized.jpg
```

#### Social media sizing

```bash
# Instagram post (max 1080x1080, maintains aspect)
imgr transform -w 1080 -h 1080 photo.jpg instagram.jpg

# Fit within specific dimensions
imgr transform -w 1500 -h 500 header.jpg twitter.jpg
```

#### Programmatic use

```bash
# Get dimensions from a script
result=$(imgr --json info photo.jpg)
width=$(echo "$result" | jq '.data.width')
height=$(echo "$result" | jq '.data.height')

# Check for errors
if echo "$result" | jq -e '.success' > /dev/null; then
  echo "Image is ${width}x${height}"
else
  echo "Error: $(echo "$result" | jq -r '.error.message')"
fi
```

## Aspect Ratio Behavior

By default, imgr maintains aspect ratio and uses dimensions as maximum bounds:

| Original | Command | Result | Notes |
|----------|---------|--------|-------|
| 1920×1080 | `-w 800` | 800×450 | Width-limited |
| 1920×1080 | `-h 600` | 1067×600 | Height-limited |
| 1920×1080 | `-w 800 -h 600` | 800×450 | Width-limited (fits within 800×600) |
| 1080×1920 | `-w 800 -h 600` | 337×600 | Height-limited (fits within 800×600) |
| 800×600 | `-w 2000 --no-enlarge` | 800×600 | No enlargement |

## Building

### Requirements

- Go 1.16 or later
- libheif (for HEIC support)
- CGO enabled (for libheif)

### Build

```bash
./build.sh
```

This creates `build/imgr` for your current platform.

### Manual Build

```bash
# Initialize module
go mod init imgr
go mod tidy

# Build
CGO_ENABLED=1 go build -o build/imgr imgr.go
```

## Testing

```bash
# Run all tests
go test -v

# Run with coverage
go test -cover

# Run specific test
go test -v -run TestLoadImageJPEG
```

Test images are in `testdata/`. The test suite includes:
- Format loading (JPEG, PNG, HEIC)
- Format conversion
- Resizing algorithms
- Aspect ratio preservation
- Dimension validation
- Info command
- JSON output
- Error handling

## Dependencies

### Build Time
- `github.com/urfave/cli/v2` - CLI framework
- `golang.org/x/image` - Image processing (draw, tiff, webp)
- `github.com/strukturag/libheif/go/heif` - HEIC support

### Runtime
- **libheif** - Only for HEIC/HEIC format support

All other formats (JPEG, PNG, GIF, TIFF, WebP) are pure Go with zero runtime dependencies.

## Resource Profile

**Binary Size:** ~4 MB  
**Memory Usage:** Scales with image size, typically 10-50 MB for common images  
**Startup Time:** <10ms  
**Dependencies:** 1 system library (libheif, HEIC only) vs 20+ for comparable tools

## What imgr includes

- Core image formats (JPEG, PNG, GIF, TIFF, WebP, HEIC)
- Resizing with high-quality interpolation
- Format conversion
- Aspect ratio preservation
- Image metadata inspection
- JSON output for programmatic use

## What imgr doesn't include

- Font rendering and text operations
- Vector graphics (SVG)
- Document formats (PDF, PostScript)
- Complex filters and effects
- RAW image formats
- Video processing

For these features, consider specialized tools like ImageMagick, FFmpeg, or dedicated libraries.

## Contributing

Contributions welcome! Please ensure:

- **Code follows Go conventions** - Use standard Go idioms and patterns.
- **Full variable names** - Avoid abbreviations except common ones like `img`.
- **Consistent spacing** - Include spaces after `(` `[` `{` and before `)` `]` `}`.
- **Error messages are complete English sentences** - Start with capital letters and end with periods.
- **Tests pass** - Run `go test -v` before submitting.
- **Keep the tool minimal** - imgr focuses on core image operations; features like text rendering, complex filters, or document processing are out of scope.

The goal is to maintain a focused, reliable tool that does image resizing and conversion well without feature creep.

## License

MIT

## Credits

Built with:
- [Go](https://golang.org/) - The Go programming language
- [urfave/cli](https://github.com/urfave/cli) - CLI framework
- [libheif](https://github.com/strukturag/libheif) - HEIF/HEIC support
- [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) - Additional image formats