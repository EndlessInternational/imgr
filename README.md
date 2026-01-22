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

**Note:** libheif is only required for HEIC/HEIF/AVIF support. All other formats work without any system dependencies.

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
Read: JPEG, PNG, GIF, TIFF, BMP, WebP, HEIF/HEIC, AVIF
Write: JPEG, PNG, GIF, TIFF, BMP

**Smart Resizing:**
- The tool maintains aspect ratio by default.
- Images fit within specified bounds when both dimensions are provided.
- The optional `--no-enlarge` flag prevents upscaling of small images.
- High-quality bilinear interpolation ensures good visual results.

**Region Extraction:**
- Extract rectangular regions using pixel coordinates.
- Useful for cropping, extracting sprites, or isolating parts of images.

## Usage

### Quick Start

```bash
# Convert JPEG to PNG
imgr transform photo.jpg photo.png

# Resize to 800px wide (maintains aspect ratio)
imgr transform -w 800 large.jpg small.jpg

# Fit within 1920x1080 (maintains aspect ratio)
imgr transform -w 1920 -h 1080 photo.jpg wallpaper.jpg

# Extract a region from an image
imgr clip --x1 100 --y1 100 --x2 500 --y2 400 photo.jpg cropped.jpg

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
- `-r, --rotate N` - Rotates the image clockwise by N degrees (90, 180, or 270).
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

# Rotate 90° clockwise
imgr transform --rotate 90 photo.jpg rotated.jpg

# Rotate and resize
imgr transform --rotate 90 -w 800 photo.jpg rotated-small.jpg

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

#### clip

Extract a rectangular region from an image.

```bash
imgr clip [options] <input> <output>
```

**Flags:**
- `--x1 N` - Left edge x coordinate (required).
- `--y1 N` - Top edge y coordinate (required).
- `--x2 N` - Right edge x coordinate (required).
- `--y2 N` - Bottom edge y coordinate (required).
- `-q, --quality N` - Sets the JPEG quality from 0 to 100 (default: 90).

**Examples:**

```bash
# Extract a 400x300 region starting at (100, 100)
imgr clip --x1 100 --y1 100 --x2 500 --y2 400 photo.jpg cropped.jpg

# Extract top-left corner (first 500x500 pixels)
imgr clip --x1 0 --y1 0 --x2 500 --y2 500 photo.jpg corner.jpg

# Extract and save as PNG
imgr clip --x1 200 --y1 150 --x2 800 --y2 600 photo.jpg region.png

# Extract with high JPEG quality
imgr clip --x1 0 --y1 0 --x2 1000 --y2 1000 -q 95 photo.jpg hq-crop.jpg
```

**Notes:**
- Coordinates are in pixels, with (0, 0) at the top-left corner.
- x2 must be greater than x1, and y2 must be greater than y1.
- Coordinates must not exceed the image dimensions.
- The output format is determined by the output file extension.

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

#### Extract regions from images

```bash
# Crop out a face or object
imgr clip --x1 200 --y1 100 --x2 600 --y2 500 photo.jpg face.jpg

# Extract multiple regions from a sprite sheet
imgr clip --x1 0 --y1 0 --x2 64 --y2 64 sprites.png sprite1.png
imgr clip --x1 64 --y1 0 --x2 128 --y2 64 sprites.png sprite2.png

# Crop and then resize
imgr clip --x1 100 --y1 100 --x2 900 --y2 700 photo.jpg cropped.jpg
imgr transform -w 400 cropped.jpg thumbnail.jpg
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
- Format loading (JPEG, PNG, HEIC, AVIF, BMP)
- Format conversion
- Resizing algorithms
- Aspect ratio preservation
- Dimension validation
- Clipping regions
- Clip coordinate validation
- Info command
- JSON output
- Error handling

## Dependencies

### Build Time
- `github.com/urfave/cli/v2` - CLI framework
- `golang.org/x/image` - Image processing (draw, tiff, webp)
- `github.com/strukturag/libheif/go/heif` - HEIC support

### Runtime
- **libheif** - Only for HEIC/HEIF/AVIF format support

All other formats (JPEG, PNG, GIF, TIFF, BMP, WebP) are pure Go with zero runtime dependencies.

## Resource Profile

**Binary Size:** ~4 MB  
**Memory Usage:** Scales with image size, typically 10-50 MB for common images  
**Startup Time:** <10ms  
**Dependencies:** 1 system library (libheif, HEIC/AVIF only) vs 20+ for comparable tools

## What imgr includes

- Core image formats (JPEG, PNG, GIF, TIFF, BMP, WebP, HEIC, AVIF)
- Resizing with high-quality interpolation
- Region extraction (clipping)
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
- **Indentation uses 2 spaces** - Use spaces only, not tabs.
- **Full variable names** - Avoid abbreviations except common ones like `err`.
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