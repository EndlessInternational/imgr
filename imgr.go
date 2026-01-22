package main

import (
  "encoding/json"
  "fmt"
  "image"
  "image/gif"
  "image/jpeg"
  "image/png"
  _ "image/gif"
  _ "image/jpeg"
  _ "image/png"
  "os"
  "path/filepath"
  "strings"

  "github.com/strukturag/libheif/go/heif"
  "github.com/urfave/cli/v2"
  "golang.org/x/image/bmp"
  "golang.org/x/image/draw"
  "golang.org/x/image/tiff"
  _ "golang.org/x/image/bmp"
  _ "golang.org/x/image/tiff"
  _ "golang.org/x/image/webp"
)

type Size struct {
  Width                 int `json:"width"`
  Height                int `json:"height"`
}

type TransformResult struct {
  InputFile             string `json:"input_file"`
  OutputFile            string `json:"output_file"`
  Format                string `json:"format"`
  OriginalSize          Size   `json:"original_size"`
  FinalSize             Size   `json:"final_size"`
  Resized               bool   `json:"resized"`
  Message               string `json:"message"`
}

type ClipResult struct {
  InputFile             string `json:"input_file"`
  OutputFile            string `json:"output_file"`
  Format                string `json:"format"`
  OriginalSize          Size   `json:"original_size"`
  ClipRegion            struct {
    X1                int `json:"x1"`
    Y1                int `json:"y1"`
    X2                int `json:"x2"`
    Y2                int `json:"y2"`
  }                     `json:"clip_region"`
  ClipSize              Size   `json:"clip_size"`
  Message               string `json:"message"`
}

type InfoResult struct {
  File                  string  `json:"file"`
  Path                  string  `json:"path"`
  Format                string  `json:"format"`
  Width                 int     `json:"width"`
  Height                int     `json:"height"`
  AspectRatio           float64 `json:"aspect_ratio"`
  HasAlpha              bool    `json:"has_alpha"`
  ColorModel            string  `json:"color_model"`
  FileSize              int64   `json:"file_size_bytes"`
  FileSizeKB            float64 `json:"file_size_kb"`
}

type ErrorResult struct {
  Message               string `json:"message"`
}

type CommandResult struct {
  Success               bool         `json:"success"`
  Data                  interface{}  `json:"data,omitempty"`
  Error                 *ErrorResult `json:"error,omitempty"`
}

func main() {
  cli.HelpFlag = &cli.BoolFlag{
    Name:  "help",
    Usage: "show help",
  }

  app := &cli.App{
    Name:             "imgr",
    Usage:            "A minimal image manipulator.",
    Description:      "A lightweight tool for resizing and converting images with low " +
                      "footprint and minimal runtime dependencies.\n" +
                      "Supports reading: JPEG, PNG, GIF, TIFF, BMP, WebP, HEIF/HEIC, AVIF.\n" +
                      "Supports writing: JPEG, PNG, GIF, TIFF, BMP.\n\n",
    Version:          "1.7.0",
    Flags: []cli.Flag{
      &cli.BoolFlag{
        Name:         "json",
        Usage:        "output results as JSON",
      },
    },
    Commands: []*cli.Command{
      {
        Name:         "transform",
        Usage:        "Resize or convert an image",
        UsageText:    "imgr transform [options] <input> <output>",
        Flags: []cli.Flag{
          &cli.IntFlag{
            Name:     "width",
            Aliases:  []string{ "w" },
            Usage:    "output width in pixels (or maximum width)",
            Value:    0,
          },
          &cli.IntFlag{
            Name:     "height",
            Aliases:  []string{ "h" },
            Usage:    "output height in pixels (or maximum height)",
            Value:    0,
          },
          &cli.IntFlag{
            Name:     "quality",
            Aliases:  []string{ "q" },
            Usage:    "JPEG quality (0-100)",
            Value:    90,
          },
          &cli.BoolFlag{
            Name:     "no-enlarge",
            Usage:    "never make image larger than source",
          },
          &cli.IntFlag{
            Name:     "rotate",
            Aliases:  []string{ "r" },
            Usage:    "rotate image clockwise (90, 180, or 270 degrees)",
            Value:    0,
          },
        },
        Action: transformImageCommand,
      },
      {
        Name:         "info",
        Usage:        "Display information about an image",
        UsageText:    "imgr info <input>",
        Action:       imageInfoCommand,
      },
      {
        Name:         "clip",
        Usage:        "Extract a rectangular region from an image",
        UsageText:    "imgr clip [options] <input> <output>",
        Flags: []cli.Flag{
          &cli.IntFlag{
            Name:       "x1",
            Usage:      "left edge x coordinate",
            Required:   true,
          },
          &cli.IntFlag{
            Name:       "y1",
            Usage:      "top edge y coordinate",
            Required:   true,
          },
          &cli.IntFlag{
            Name:       "x2",
            Usage:      "right edge x coordinate",
            Required:   true,
          },
          &cli.IntFlag{
            Name:       "y2",
            Usage:      "bottom edge y coordinate",
            Required:   true,
          },
          &cli.IntFlag{
            Name:     "quality",
            Aliases:  []string{ "q" },
            Usage:    "JPEG quality (0-100)",
            Value:    90,
          },
        },
        Action: clipImageCommand,
      },
    },
  }

  if err := app.Run( os.Args ); err != nil {
    // the error has already been output by the command, just exit
    os.Exit( 1 )
  }
}

func outputError( message string, useJSON bool ) {
  result := CommandResult{
    Success: false,
    Error: &ErrorResult{
      Message: message,
    },
  }

  if useJSON {
    json.NewEncoder( os.Stdout ).Encode( result )
  } else {
    fmt.Fprintf( os.Stderr, "Error: %s\n", message )
  }
}

func outputSuccess( data interface{}, useJSON bool ) {
  if useJSON {
    result := CommandResult{
      Success: true,
      Data:    data,
    }
    json.NewEncoder( os.Stdout ).Encode( result )
  }
}

func loadImage( path string ) ( image.Image, string, error ) {
  extension := strings.ToLower( filepath.Ext( path ) )

  if extension == ".heic" || extension == ".heif" || extension == ".avif" {
    return decodeHeif( path )
  }

  inputFile, err := os.Open( path )
  if err != nil {
    return nil, "", err
  }
  defer inputFile.Close()

  return image.Decode( inputFile )
}

func decodeHeif( path string ) ( image.Image, string, error ) {
  heifContext, err := heif.NewContext()
  if err != nil {
    return nil, "", fmt.Errorf( "The HEIF context could not be created: %w", err )
  }

  err = heifContext.ReadFromFile( path )
  if err != nil {
    return nil, "", fmt.Errorf( "The HEIF file could not be read: %w", err )
  }

  handle, err := heifContext.GetPrimaryImageHandle()
  if err != nil {
    return nil, "", fmt.Errorf( "The primary image could not be retrieved: %w", err )
  }

  img, err := handle.DecodeImage( heif.ColorspaceUndefined, heif.ChromaUndefined, nil )
  if err != nil {
    return nil, "", fmt.Errorf( "The HEIF image could not be decoded: %w", err )
  }

  goImage, err := img.GetImage()
  if err != nil {
    return nil, "", fmt.Errorf( "The HEIF image could not be converted: %w", err )
  }

  return goImage, "heif", nil
}

func rotateImage( img image.Image, degrees int ) image.Image {
  bounds := img.Bounds()
  width := bounds.Dx()
  height := bounds.Dy()

  switch degrees {
  case 90:
    rotated := image.NewRGBA( image.Rect( 0, 0, height, width ) )
    for y := 0; y < height; y++ {
      for x := 0; x < width; x++ {
        rotated.Set( height-1-y, x, img.At( x+bounds.Min.X, y+bounds.Min.Y ) )
      }
    }
    return rotated

  case 180:
    rotated := image.NewRGBA( image.Rect( 0, 0, width, height ) )
    for y := 0; y < height; y++ {
      for x := 0; x < width; x++ {
        rotated.Set( width-1-x, height-1-y, img.At( x+bounds.Min.X, y+bounds.Min.Y ) )
      }
    }
    return rotated

  case 270:
    rotated := image.NewRGBA( image.Rect( 0, 0, height, width ) )
    for y := 0; y < height; y++ {
      for x := 0; x < width; x++ {
        rotated.Set( y, width-1-x, img.At( x+bounds.Min.X, y+bounds.Min.Y ) )
      }
    }
    return rotated

  default:
    return img
  }
}

func transformImageCommand( context *cli.Context ) error {
  useJSON := context.Bool( "json" )
  result, err := transformImage( context )

  if err != nil {
    outputError( err.Error(), useJSON )
    return err
  }

  if useJSON {
    outputSuccess( result, useJSON )
  } else {
    fmt.Println( result.Message )
    fmt.Printf( "✓ Saved to %s\n", result.OutputFile )
  }

  return nil
}

func transformImage( context *cli.Context ) ( *TransformResult, error ) {
  if context.NArg() != 2 {
    return nil, fmt.Errorf( "Expected 2 arguments (input and output), but got %d.", context.NArg() )
  }

  inputPath := context.Args().Get( 0 )
  outputPath := context.Args().Get( 1 )
  maxWidth := context.Int( "width" )
  maxHeight := context.Int( "height" )
  quality := context.Int( "quality" )
  noEnlarge := context.Bool( "no-enlarge" )
  rotate := context.Int( "rotate" )

  if rotate != 0 && rotate != 90 && rotate != 180 && rotate != 270 {
    return nil, fmt.Errorf( "Rotation must be 0, 90, 180, or 270 degrees, but got %d.", rotate )
  }

  if maxWidth < 0 {
    return nil, fmt.Errorf( "Width cannot be negative, but got %d.", maxWidth )
  }

  if maxHeight < 0 {
    return nil, fmt.Errorf( "Height cannot be negative, but got %d.", maxHeight )
  }

  if quality < 0 || quality > 100 {
    return nil, fmt.Errorf( "Quality must be between 0 and 100, but got %d.", quality )
  }

  sourceImage, format, err := loadImage( inputPath )
  if err != nil {
    return nil, fmt.Errorf( "The image file %s could not be decoded (possibly corrupt or unsupported format): %w",
      inputPath, err )
  }

  if sourceImage == nil {
    return nil, fmt.Errorf( "The decoded image from %s is invalid.", inputPath )
  }

  // apply rotation before resizing
  if rotate != 0 {
    sourceImage = rotateImage( sourceImage, rotate )
  }

  bounds := sourceImage.Bounds()
  originalWidth := bounds.Dx()
  originalHeight := bounds.Dy()

  if originalWidth <= 0 || originalHeight <= 0 {
    return nil, fmt.Errorf( "The image %s has invalid dimensions: %dx%d.",
      inputPath, originalWidth, originalHeight )
  }

  const maxDimension = 65535
  if originalWidth > maxDimension || originalHeight > maxDimension {
    return nil, fmt.Errorf( "The image %s is too large: %dx%d (maximum dimension is %d).",
      inputPath, originalWidth, originalHeight, maxDimension )
  }

  var destinationImage image.Image
  var targetWidth, targetHeight int
  var resized bool
  var message string

  if maxWidth == 0 && maxHeight == 0 {
    message = fmt.Sprintf( "Converting %s [%s] %dx%d (no resize)",
      filepath.Base( inputPath ),
      format,
      originalWidth,
      originalHeight,
    )
    destinationImage = sourceImage
    targetWidth = originalWidth
    targetHeight = originalHeight
    resized = false
  } else {
    targetWidth = maxWidth
    targetHeight = maxHeight

    if maxWidth == 0 && maxHeight > 0 {
      aspectRatio := float64( originalWidth ) / float64( originalHeight )
      targetWidth = int( float64( targetHeight ) * aspectRatio + 0.5 )
    } else if maxHeight == 0 && maxWidth > 0 {
      aspectRatio := float64( originalHeight ) / float64( originalWidth )
      targetHeight = int( float64( targetWidth ) * aspectRatio + 0.5 )
    } else {
      originalAspect := float64( originalWidth ) / float64( originalHeight )
      targetAspect := float64( maxWidth ) / float64( maxHeight )

      if originalAspect > targetAspect {
        targetWidth = maxWidth
        targetHeight = int( float64( maxWidth ) / originalAspect + 0.5 )
      } else {
        targetHeight = maxHeight
        targetWidth = int( float64( maxHeight ) * originalAspect + 0.5 )
      }
    }

    if noEnlarge {
      if targetWidth > originalWidth || targetHeight > originalHeight {
        targetWidth = originalWidth
        targetHeight = originalHeight
      }
    }

    if targetWidth <= 0 || targetHeight <= 0 {
      return nil, fmt.Errorf( "Calculated target dimensions are invalid: %dx%d.",
        targetWidth, targetHeight )
    }

    if targetWidth > maxDimension || targetHeight > maxDimension {
      return nil, fmt.Errorf( "Target dimensions %dx%d exceed maximum dimension of %d.",
        targetWidth, targetHeight, maxDimension )
    }

    if targetWidth == originalWidth && targetHeight == originalHeight {
      message = fmt.Sprintf( "Converting %s [%s] %dx%d (no resize needed)",
        filepath.Base( inputPath ),
        format,
        originalWidth,
        originalHeight,
      )
      destinationImage = sourceImage
      resized = false
    } else {
      resizeMode := "maintaining aspect ratio"
      if maxWidth > 0 && maxHeight > 0 {
        resizeMode = fmt.Sprintf( "fit within %dx%d", maxWidth, maxHeight )
      }
      if noEnlarge {
        resizeMode += ", no enlargement"
      }

      message = fmt.Sprintf( "Resizing %s [%s] from %dx%d to %dx%d (%s)",
        filepath.Base( inputPath ),
        format,
        originalWidth,
        originalHeight,
        targetWidth,
        targetHeight,
        resizeMode,
      )

      resizedImage := image.NewRGBA( image.Rect( 0, 0, targetWidth, targetHeight ) )

      draw.BiLinear.Scale(
        resizedImage,
        resizedImage.Bounds(),
        sourceImage,
        sourceImage.Bounds(),
        draw.Over,
        nil,
      )

      destinationImage = resizedImage
      resized = true
    }
  }

  outputExtension := strings.ToLower( filepath.Ext( outputPath ) )
  err = encodeOutput( outputPath, outputExtension, destinationImage, quality, format )
  if err != nil {
    return nil, fmt.Errorf( "The output file %s could not be written: %w", outputPath, err )
  }

  return &TransformResult{
    InputFile:  inputPath,
    OutputFile: outputPath,
    Format:     format,
    OriginalSize: Size{
      Width:  originalWidth,
      Height: originalHeight,
    },
    FinalSize: Size{
      Width:  targetWidth,
      Height: targetHeight,
    },
    Resized: resized,
    Message: message,
  }, nil
}

func imageInfoCommand( context *cli.Context ) error {
  useJSON := context.Bool( "json" )
  result, err := imageInfo( context )

  if err != nil {
    outputError( err.Error(), useJSON )
    return err
  }

  if useJSON {
    outputSuccess( result, useJSON )
  } else {
    fmt.Printf( "File:         %s\n", result.File )
    fmt.Printf( "Path:         %s\n", result.Path )
    fmt.Printf( "Format:       %s\n", strings.ToUpper( result.Format ) )
    fmt.Printf( "Dimensions:   %d × %d pixels\n", result.Width, result.Height )
    fmt.Printf( "Aspect Ratio: %.2f:1\n", result.AspectRatio )
    fmt.Printf( "Transparency: %v\n", result.HasAlpha )
    fmt.Printf( "Color Model:  %s\n", result.ColorModel )
    fmt.Printf( "File Size:    %d bytes (%.2f KB)\n", result.FileSize, result.FileSizeKB )
  }

  return nil
}

func imageInfo( context *cli.Context ) ( *InfoResult, error ) {
  if context.NArg() != 1 {
    return nil, fmt.Errorf( "Expected 1 argument (input file), but got %d.", context.NArg() )
  }

  inputPath := context.Args().Get( 0 )

  fileInfo, err := os.Stat( inputPath )
  if err != nil {
    return nil, fmt.Errorf( "The file %s could not be accessed: %w", inputPath, err )
  }

  if fileInfo.Size() == 0 {
    return nil, fmt.Errorf( "The file %s is empty.", inputPath )
  }

  sourceImage, format, err := loadImage( inputPath )
  if err != nil {
    return nil, fmt.Errorf( "The image file %s could not be decoded (possibly corrupt or unsupported format): %w",
      inputPath,
      err )
  }

  if sourceImage == nil {
    return nil, fmt.Errorf( "The decoded image from %s is invalid.", inputPath )
  }

  bounds := sourceImage.Bounds()
  width := bounds.Dx()
  height := bounds.Dy()

  if width <= 0 || height <= 0 {
    return nil, fmt.Errorf( "The image %s has invalid dimensions: %dx%d.", inputPath, width, height )
  }

  hasAlpha := false
  switch sourceImage.( type ) {
  case *image.NRGBA, *image.NRGBA64, *image.RGBA, *image.RGBA64, *image.Alpha, *image.Alpha16:
    hasAlpha = true
  }

  colorModelName := fmt.Sprintf( "%T", sourceImage )
  colorModelName = strings.TrimPrefix( colorModelName, "*image." )

  aspectRatio := float64( width ) / float64( height )

  return &InfoResult{
    File:        filepath.Base( inputPath ),
    Path:        inputPath,
    Format:      format,
    Width:       width,
    Height:      height,
    AspectRatio: aspectRatio,
    HasAlpha:    hasAlpha,
    ColorModel:  colorModelName,
    FileSize:    fileInfo.Size(),
    FileSizeKB:  float64( fileInfo.Size() ) / 1024.0,
  }, nil
}

func clipImageCommand( context *cli.Context ) error {
  useJSON := context.Bool( "json" )
  result, err := clipImage( context )

  if err != nil {
    outputError( err.Error(), useJSON )
    return err
  }

  if useJSON {
    outputSuccess( result, useJSON )
  } else {
    fmt.Println( result.Message )
    fmt.Printf( "✓ Saved to %s\n", result.OutputFile )
  }

  return nil
}

func clipImage( context *cli.Context ) ( *ClipResult, error ) {
  if context.NArg() != 2 {
    return nil, fmt.Errorf( "Expected 2 arguments ( input and output ), but got %d.", context.NArg() )
  }

  inputPath := context.Args().Get( 0 )
  outputPath := context.Args().Get( 1 )
  x1 := context.Int( "x1" )
  y1 := context.Int( "y1" )
  x2 := context.Int( "x2" )
  y2 := context.Int( "y2" )
  quality := context.Int( "quality" )

  if quality < 0 || quality > 100 {
    return nil, fmt.Errorf( "Quality must be between 0 and 100, but got %d.", quality )
  }

  if x1 < 0 || y1 < 0 || x2 < 0 || y2 < 0 {
    return nil, fmt.Errorf( "Coordinates cannot be negative." )
  }

  if x2 <= x1 {
    return nil, fmt.Errorf( "x2 must be greater than x1 ( got x1=%d, x2=%d ).", x1, x2 )
  }

  if y2 <= y1 {
    return nil, fmt.Errorf( "y2 must be greater than y1 ( got y1=%d, y2=%d ).", y1, y2 )
  }

  sourceImage, format, err := loadImage( inputPath )
  if err != nil {
    return nil, fmt.Errorf(
      "The image file %s could not be decoded ( possibly corrupt or unsupported format ): %w",
      inputPath, err )
  }

  if sourceImage == nil {
    return nil, fmt.Errorf( "The decoded image from %s is invalid.", inputPath )
  }

  bounds := sourceImage.Bounds()
  originalWidth := bounds.Dx()
  originalHeight := bounds.Dy()

  if originalWidth <= 0 || originalHeight <= 0 {
    return nil, fmt.Errorf( "The image %s has invalid dimensions: %dx%d.",
      inputPath, originalWidth, originalHeight )
  }

  if x2 > originalWidth {
    return nil, fmt.Errorf( "The x2 coordinate ( %d ) exceeds the image width ( %d ).", x2, originalWidth )
  }

  if y2 > originalHeight {
    return nil, fmt.Errorf( "The y2 coordinate ( %d ) exceeds the image height ( %d ).", y2, originalHeight )
  }

  clipWidth := x2 - x1
  clipHeight := y2 - y1

  // create the clipped image by drawing the source region onto a new image
  clippedImage := image.NewRGBA( image.Rect( 0, 0, clipWidth, clipHeight ) )
  sourceRect := image.Rect( x1, y1, x2, y2 )

  draw.Draw(
    clippedImage,
    clippedImage.Bounds(),
    sourceImage,
    sourceRect.Min,
    draw.Src,
  )

  message := fmt.Sprintf( "Clipping %s [%s] region ( %d,%d )-( %d,%d ) -> %dx%d",
    filepath.Base( inputPath ),
    format,
    x1, y1, x2, y2,
    clipWidth, clipHeight,
  )

  outputExtension := strings.ToLower( filepath.Ext( outputPath ) )
  err = encodeOutput( outputPath, outputExtension, clippedImage, quality, format )
  if err != nil {
    return nil, fmt.Errorf( "The output file %s could not be written: %w", outputPath, err )
  }

  result := &ClipResult{
    InputFile:    inputPath,
    OutputFile:   outputPath,
    Format:       format,
    OriginalSize: Size{ Width: originalWidth, Height: originalHeight },
    ClipSize:     Size{ Width: clipWidth, Height: clipHeight },
    Message:      message,
  }
  result.ClipRegion.X1 = x1
  result.ClipRegion.Y1 = y1
  result.ClipRegion.X2 = x2
  result.ClipRegion.Y2 = y2

  return result, nil
}

func encodeOutput( path string, extension string, img image.Image, quality int, inputFormat string ) error {
  outputFile, err := os.Create( path )
  if err != nil {
    return fmt.Errorf( "The output file %s could not be created: %w", path, err )
  }

  // for unknown extensions, use input format ( fall back to jpeg for formats we can't write )
  supportedExtensions := map[ string ]bool{
    ".png": true, ".gif": true, ".jpg": true, ".jpeg": true, ".tif": true, ".tiff": true, ".bmp": true,
  }

  effectiveExtension := extension
  if !supportedExtensions[ extension ] {
    switch inputFormat {
    case "png":
      effectiveExtension = ".png"
    case "gif":
      effectiveExtension = ".gif"
    case "tiff":
      effectiveExtension = ".tiff"
    case "bmp":
      effectiveExtension = ".bmp"
    default:
      effectiveExtension = ".jpeg"
    }
  }

  switch effectiveExtension {
  case ".png":
    err = png.Encode( outputFile, img )
  case ".gif":
    err = gif.Encode( outputFile, img, nil )
  case ".jpg", ".jpeg":
    options := &jpeg.Options{ Quality: quality }
    err = jpeg.Encode( outputFile, img, options )
  case ".tif", ".tiff":
    err = tiff.Encode( outputFile, img, &tiff.Options{ Compression: tiff.Deflate } )
  case ".bmp":
    err = bmp.Encode( outputFile, img )
  }

  if err != nil {
    outputFile.Close()
    os.Remove( path )
    return fmt.Errorf( "The image could not be encoded as %s: %w", effectiveExtension, err )
  }

  return outputFile.Close()
}