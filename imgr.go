package main

import (
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
  "golang.org/x/image/draw"
  "golang.org/x/image/tiff"
  _ "golang.org/x/image/tiff"
  _ "golang.org/x/image/webp"
)

func main() {
  cli.HelpFlag = &cli.BoolFlag{
    Name:  "help",
    Usage: "show help",
  }

  app := &cli.App{
    Name:               "imgr",
    Usage:              "A minimal image manipulator.",
    Description:        "A lightweight tool for resizing and converting images with low " +
                        "footprint and minimal runtime dependencies.\n" +
                        "Supports reading: JPEG, PNG, GIF, TIFF, WebP, HEIF/HEIC.\n" +
                        "Supports writing: JPEG, PNG, GIF, TIFF.\n\n",
    Version:            "1.2.0",
    Commands: []*cli.Command{
      {
        Name:           "transform",
        Usage:          "Resize or convert an image",
        UsageText:      "imgr transform [options] <input> <output>",
        Flags: []cli.Flag{
          &cli.IntFlag{
            Name:       "width",
            Aliases:    []string{ "w" },
            Usage:      "output width in pixels (or maximum width)",
            Value:      0,
          },
          &cli.IntFlag{
            Name:       "height",
            Aliases:    []string{ "h" },
            Usage:      "output height in pixels (or maximum height)",
            Value:      0,
          },
          &cli.IntFlag{
            Name:       "quality",
            Aliases:    []string{ "q" },
            Usage:      "JPEG quality (0-100)",
            Value:      90,
          },
          &cli.BoolFlag{
            Name:       "no-enlarge",
            Usage:      "never make image larger than source",
          },
        },
        Action: transformImage,
      },
      {
        Name:           "info",
        Usage:          "Display information about an image",
        UsageText:      "imgr info <input>",
        Action:         imageInfo,
      },
    },
    Action: defaultAction,
  }

  if err := app.Run( os.Args ); err != nil {
    fmt.Fprintf( os.Stderr, "Error: %v\n", err )
    os.Exit( 1 )
  }
}

func defaultAction( context *cli.Context ) error {
  args := context.Args().Slice()
  
  if len( args ) == 0 {
    return cli.ShowAppHelp( context )
  }

  newArgs := append( []string{ os.Args[0], "transform" }, args... )
  return context.App.Run( newArgs )
}

func loadImage( path string ) ( image.Image, string, error ) {
  extension := strings.ToLower( filepath.Ext( path ) )
  
  if extension == ".heic" || extension == ".heif" {
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
    return nil, "", fmt.Errorf( "Failed to create HEIF context: %w", err )
  }

  err = heifContext.ReadFromFile( path )
  if err != nil {
    return nil, "", fmt.Errorf( "Failed to read HEIF file: %w", err )
  }

  handle, err := heifContext.GetPrimaryImageHandle()
  if err != nil {
    return nil, "", fmt.Errorf( "Failed to get primary image: %w", err )
  }

  img, err := handle.DecodeImage( heif.ColorspaceUndefined, heif.ChromaUndefined, nil )
  if err != nil {
    return nil, "", fmt.Errorf( "Failed to decode HEIF image: %w", err )
  }

  goImage, err := img.GetImage()
  if err != nil {
    return nil, "", fmt.Errorf( "Failed to convert HEIF to usable format: %w", err )
  }

  return goImage, "heif", nil
}

func transformImage( context *cli.Context ) error {
  if context.NArg() != 2 {
    return fmt.Errorf( "Expected 2 arguments (input and output), but got %d.", context.NArg() )
  }

  inputPath := context.Args().Get( 0 )
  outputPath := context.Args().Get( 1 )
  maxWidth := context.Int( "width" )
  maxHeight := context.Int( "height" )
  quality := context.Int( "quality" )
  noEnlarge := context.Bool( "no-enlarge" )

  if maxWidth < 0 {
    return fmt.Errorf( "Width cannot be negative, but got %d.", maxWidth )
  }

  if maxHeight < 0 {
    return fmt.Errorf( "Height cannot be negative, but got %d.", maxHeight )
  }

  if quality < 0 || quality > 100 {
    return fmt.Errorf( "Quality must be between 0 and 100, but got %d.", quality )
  }

  sourceImage, format, err := loadImage( inputPath )
  if err != nil {
    return fmt.Errorf( "The image file %s could not be decoded (possibly corrupt or " + 
                         "unsupported format): %w", 
                       inputPath, err )
  }

  if sourceImage == nil {
    return fmt.Errorf( "The decoded image from %s is invalid.", inputPath )
  }

  bounds := sourceImage.Bounds()
  originalWidth := bounds.Dx()
  originalHeight := bounds.Dy()

  if originalWidth <= 0 || originalHeight <= 0 {
    return fmt.Errorf( "The image %s has invalid dimensions: %dx%d.", 
                       inputPath, originalWidth, originalHeight )
  }

  const maxDimension = 65535
  if originalWidth > maxDimension || originalHeight > maxDimension {
    return fmt.Errorf( "The image %s is too large: %dx%d (maximum dimension is %d).", 
                       inputPath, originalWidth, originalHeight, maxDimension )
  }

  var destinationImage image.Image

  if maxWidth == 0 && maxHeight == 0 {
    fmt.Printf( 
      "Converting %s [%s] %dx%d (no resize)\n",
      filepath.Base( inputPath ),
      format,
      originalWidth,
      originalHeight,
    )
    destinationImage = sourceImage
  } else {
    targetWidth := maxWidth
    targetHeight := maxHeight

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
      return fmt.Errorf( "Calculated target dimensions are invalid: %dx%d.", 
                         targetWidth, targetHeight )
    }

    if targetWidth > maxDimension || targetHeight > maxDimension {
      return fmt.Errorf( "Target dimensions %dx%d exceed maximum dimension of %d.", 
                         targetWidth, targetHeight, maxDimension )
    }

    if targetWidth == originalWidth && targetHeight == originalHeight {
      fmt.Printf( "Converting %s [%s] %dx%d (no resize needed)\n",
        filepath.Base( inputPath ),
        format,
        originalWidth,
        originalHeight,
      )
      destinationImage = sourceImage
    } else {
      resizeMode := "maintaining aspect ratio"
      if maxWidth > 0 && maxHeight > 0 {
        resizeMode = fmt.Sprintf( "fit within %dx%d", maxWidth, maxHeight )
      }
      if noEnlarge {
        resizeMode += ", no enlargement"
      }

      fmt.Printf( "Resizing %s [%s] from %dx%d to %dx%d (%s)\n",
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
    }
  }

  outputExtension := strings.ToLower( filepath.Ext( outputPath ) )
  err = encodeOutput( outputPath, outputExtension, destinationImage, quality )
  if err != nil {
    return fmt.Errorf( "The output file %s could not be written: %w", outputPath, err )
  }

  fmt.Printf( "✓ Saved to %s\n", outputPath )
  return nil
}

func imageInfo( context *cli.Context ) error {
  if context.NArg() != 1 {
    return fmt.Errorf( "Expected 1 argument (input file), but got %d.", context.NArg() )
  }

  inputPath := context.Args().Get( 0 )

  fileInfo, err := os.Stat( inputPath )
  if err != nil {
    return fmt.Errorf( "The file %s could not be accessed: %w", inputPath, err )
  }

  if fileInfo.Size() == 0 {
    return fmt.Errorf( "The file %s is empty.", inputPath )
  }

  sourceImage, format, err := loadImage( inputPath )
  if err != nil {
    return fmt.Errorf( "The image file %s could not be decoded (possibly " +
                         "corrupt or unsupported format): %w", 
                       inputPath, 
                       err )
  }

  if sourceImage == nil {
    return fmt.Errorf( "The decoded image from %s is invalid.", inputPath )
  }

  bounds := sourceImage.Bounds()
  width := bounds.Dx()
  height := bounds.Dy()

  if width <= 0 || height <= 0 {
    return fmt.Errorf( "The image %s has invalid dimensions: %dx%d.", inputPath, width, height )
  }

  hasAlpha := false
  switch sourceImage.( type ) {
  case *image.NRGBA, *image.NRGBA64, *image.RGBA, *image.RGBA64, *image.Alpha, *image.Alpha16:
    hasAlpha = true
  }

  colorModelName := fmt.Sprintf( "%T", sourceImage )
  colorModelName = strings.TrimPrefix( colorModelName, "*image." )

  aspectRatio := float64( width ) / float64( height )

  fmt.Printf( "File:         %s\n", filepath.Base( inputPath ) )
  fmt.Printf( "Path:         %s\n", inputPath )
  fmt.Printf( "Format:       %s\n", strings.ToUpper( format ) )
  fmt.Printf( "Dimensions:   %d × %d pixels\n", width, height )
  fmt.Printf( "Aspect Ratio: %.2f:1\n", aspectRatio )
  fmt.Printf( "Transparency: %v\n", hasAlpha )
  fmt.Printf( "Color Model:  %s\n", colorModelName )
  fmt.Printf( "File Size:    %d bytes (%.2f KB)\n", 
              fileInfo.Size(), float64( fileInfo.Size() ) / 1024.0 )

  return nil
}

func encodeOutput( path string, extension string, img image.Image, quality int ) error {
  outputFile, err := os.Create( path )
  if err != nil {
    return fmt.Errorf( "The output file %s could not be created: %w", path, err )
  }
  defer outputFile.Close()

  switch extension {
  case ".png":
    err = png.Encode( outputFile, img )
  case ".gif":
    err = gif.Encode( outputFile, img, nil )
  case ".jpg", ".jpeg":
    options := &jpeg.Options{ Quality: quality }
    err = jpeg.Encode( outputFile, img, options )
  case ".tif", ".tiff":
    err = tiff.Encode( outputFile, img, &tiff.Options{ Compression: tiff.Deflate } )
  default:
    options := &jpeg.Options{ Quality: quality }
    err = jpeg.Encode( outputFile, img, options )
  }

  if err != nil {
    return fmt.Errorf( "Failed to encode image as %s: %w", extension, err )
  }

  return nil
}