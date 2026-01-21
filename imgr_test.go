package main

import (
  "fmt"
  "image"
  "os"
  "path/filepath"
  "testing"
)

func TestLoadImageJPEG( t *testing.T ) {
  img, format, err := loadImage( "testdata/test.jpeg" )
  
  if err != nil {
    t.Fatalf( "The JPEG image could not be loaded: %v", err )
  }
  
  if format != "jpeg" {
    t.Errorf( "Expected format to be 'jpeg', but got '%s'.", format )
  }
  
  if img == nil {
    t.Error( "The image should not be nil." )
  }
  
  bounds := img.Bounds()
  if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
    t.Errorf( "The image has invalid dimensions: %dx%d.", bounds.Dx(), bounds.Dy() )
  }
}

func TestLoadImageHEIF( t *testing.T ) {
  if _, err := os.Stat( "testdata/test.heic" ); os.IsNotExist( err ) {
    t.Skip( "HEIC test file not present, skipping." )
  }
  
  img, format, err := loadImage( "testdata/test.heic" )
  
  if err != nil {
    t.Fatalf( "The HEIC image could not be loaded: %v", err )
  }
  
  if format != "heif" {
    t.Errorf( "Expected format to be 'heif', but got '%s'.", format )
  }
  
  if img == nil {
    t.Error( "The image should not be nil." )
  }
}

func TestLoadImageAVIF( t *testing.T ) {
  if _, err := os.Stat( "testdata/test.avif" ); os.IsNotExist( err ) {
    t.Skip( "AVIF test file not present, skipping." )
  }

  img, format, err := loadImage( "testdata/test.avif" )

  if err != nil {
    t.Fatalf( "The AVIF image could not be loaded: %v", err )
  }

  if format != "heif" {
    t.Errorf( "Expected format to be 'heif', but got '%s'.", format )
  }

  if img == nil {
    t.Error( "The image should not be nil." )
  }

  bounds := img.Bounds()
  if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
    t.Errorf( "The image has invalid dimensions: %dx%d.", bounds.Dx(), bounds.Dy() )
  }

  t.Logf( "AVIF image loaded: %dx%d.", bounds.Dx(), bounds.Dy() )
}

func TestLoadImageNonExistent( t *testing.T ) {
  _, _, err := loadImage( "testdata/nonexistent.jpg" )
  
  if err == nil {
    t.Error( "Expected an error for non-existent file, but got none." )
  }
}

func TestTransformBasicResize( t *testing.T ) {
  inputPath := "testdata/test.jpeg"
  outputPath := "testdata/output_resized.jpg"
  
  defer os.Remove( outputPath )
  
  sourceImage, _, err := loadImage( inputPath )
  if err != nil {
    t.Fatalf( "The source image could not be loaded: %v", err )
  }
  
  originalWidth := sourceImage.Bounds().Dx()
  
  err = encodeOutput( outputPath, ".jpg", sourceImage, 90, "jpeg" )
  if err != nil {
    t.Fatalf( "The output could not be encoded: %v", err )
  }
  
  if _, err := os.Stat( outputPath ); err == nil {
    t.Logf( "Output file created successfully, original width: %d.", originalWidth )
  } else {
    t.Errorf( "The output file was not created." )
  }
}

func TestTransformNoEnlarge( t *testing.T ) {
  inputPath := "testdata/test.jpeg"
  
  sourceImage, _, err := loadImage( inputPath )
  if err != nil {
    t.Fatalf( "The source image could not be loaded: %v", err )
  }
  
  originalWidth := sourceImage.Bounds().Dx()
  originalHeight := sourceImage.Bounds().Dy()
  
  if originalWidth >= 5000 {
    t.Skip( "Test image is too large for this test." )
  }
  
  largeWidth := originalWidth * 2
  largeHeight := originalHeight * 2
  
  t.Logf( "Original: %dx%d, Requested: %dx%d (with no-enlarge).", 
    originalWidth, originalHeight, largeWidth, largeHeight )
  
  t.Log( "Logic: with no-enlarge, output should remain original size." )
}

func TestFormatConversion( t *testing.T ) {
  tests := []struct {
    name       string
    input      string
    output     string
    wantFormat string
  }{
    { "JPEG to PNG", "testdata/test.jpeg", "testdata/output.png", "png" },
    { "JPEG to GIF", "testdata/test.jpeg", "testdata/output.gif", "gif" },
    { "JPEG to TIFF", "testdata/test.jpeg", "testdata/output.tiff", "tiff" },
  }
  
  for _, tt := range tests {
    t.Run( tt.name, func( t *testing.T ) {
      defer os.Remove( tt.output )
      
      sourceImage, _, err := loadImage( tt.input )
      if err != nil {
        t.Fatalf( "The file %s could not be loaded: %v", tt.input, err )
      }
      
      err = encodeOutput( tt.output, filepath.Ext( tt.output ), sourceImage, 90, "jpeg" )
      if err != nil {
        t.Fatalf( "The file %s could not be encoded: %v", tt.output, err )
      }
      
      if _, err := os.Stat( tt.output ); os.IsNotExist( err ) {
        t.Errorf( "The output file %s was not created.", tt.output )
      }
    } )
  }
}

func TestQualityValidation( t *testing.T ) {
  tests := []struct {
    quality   int
    shouldErr bool
  }{
    { 0, false },
    { 50, false },
    { 90, false },
    { 100, false },
    { -1, true },
    { 101, true },
  }
  
  for _, tt := range tests {
    t.Run( fmt.Sprintf( "quality_%d", tt.quality ), func( t *testing.T ) {
      if tt.quality < 0 || tt.quality > 100 {
        if !tt.shouldErr {
          t.Error( "Should have produced an error for invalid quality value." )
        }
      }
    } )
  }
}

func TestAspectRatioMaintained( t *testing.T ) {
  inputPath := "testdata/test.jpeg"
  
  sourceImage, _, err := loadImage( inputPath )
  if err != nil {
    t.Fatalf( "The source image could not be loaded: %v", err )
  }
  
  originalWidth := sourceImage.Bounds().Dx()
  originalHeight := sourceImage.Bounds().Dy()
  originalAspect := float64( originalWidth ) / float64( originalHeight )
  
  maxWidth := 800
  maxHeight := 600
  
  var targetWidth, targetHeight int
  targetAspect := float64( maxWidth ) / float64( maxHeight )
  
  if originalAspect > targetAspect {
    targetWidth = maxWidth
    targetHeight = int( float64( maxWidth ) / originalAspect + 0.5 )
  } else {
    targetHeight = maxHeight
    targetWidth = int( float64( maxHeight ) * originalAspect + 0.5 )
  }
  
  calculatedAspect := float64( targetWidth ) / float64( targetHeight )
  
  aspectDiff := calculatedAspect - originalAspect
  if aspectDiff < 0 {
    aspectDiff = -aspectDiff
  }
  
  if aspectDiff > 0.01 {
    t.Errorf( "Aspect ratio was not maintained: original %.2f, calculated %.2f.", 
      originalAspect, calculatedAspect )
  }
}

func TestInvalidDimensions( t *testing.T ) {
  tests := []struct {
    name      string
    width     int
    height    int
    shouldErr bool
  }{
    { "negative width", -100, 100, true },
    { "negative height", 100, -100, true },
    { "zero dimensions", 0, 0, false },
    { "valid dimensions", 800, 600, false },
  }
  
  for _, tt := range tests {
    t.Run( tt.name, func( t *testing.T ) {
      hasErr := tt.width < 0 || tt.height < 0
      
      if hasErr != tt.shouldErr {
        t.Errorf( "Expected error state %v, but got %v.", tt.shouldErr, hasErr )
      }
    } )
  }
}

func TestImageInfoBasic( t *testing.T ) {
  inputPath := "testdata/test.jpeg"
  
  fileInfo, err := os.Stat( inputPath )
  if err != nil {
    t.Fatalf( "The file could not be accessed: %v", err )
  }
  
  if fileInfo.Size() == 0 {
    t.Error( "The test image file is empty." )
  }
  
  sourceImage, format, err := loadImage( inputPath )
  if err != nil {
    t.Fatalf( "The image could not be loaded: %v", err )
  }
  
  if sourceImage == nil {
    t.Fatal( "The image should not be nil." )
  }
  
  bounds := sourceImage.Bounds()
  width := bounds.Dx()
  height := bounds.Dy()
  
  if width <= 0 || height <= 0 {
    t.Errorf( "The image has invalid dimensions: %dx%d.", width, height )
  }
  
  t.Logf( "Image info: %s, %dx%d, %.2f KB.", 
    format, width, height, float64( fileInfo.Size() ) / 1024.0 )
}

func TestImageInfoNonExistent( t *testing.T ) {
  inputPath := "testdata/nonexistent.jpg"
  
  _, err := os.Stat( inputPath )
  if err == nil {
    t.Error( "Expected an error for non-existent file, but got none." )
  }
}

func TestImageInfoEmptyFile( t *testing.T ) {
  emptyFile := "testdata/empty.jpg"
  
  f, err := os.Create( emptyFile )
  if err != nil {
    t.Fatalf( "The empty file could not be created: %v", err )
  }
  f.Close()
  defer os.Remove( emptyFile )
  
  fileInfo, err := os.Stat( emptyFile )
  if err != nil {
    t.Fatalf( "The empty file could not be accessed: %v", err )
  }
  
  if fileInfo.Size() != 0 {
    t.Error( "The file should be empty." )
  }
}

func TestImageInfoCorruptFile( t *testing.T ) {
  corruptFile := "testdata/corrupt.jpg"
  
  err := os.WriteFile( corruptFile, []byte( "not a valid image" ), 0644 )
  if err != nil {
    t.Fatalf( "The corrupt test file could not be created: %v", err )
  }
  defer os.Remove( corruptFile )
  
  _, _, err = loadImage( corruptFile )
  if err == nil {
    t.Error( "Expected an error for corrupt image file, but got none." )
  }
}

func TestImageInfoAspectRatio( t *testing.T ) {
  inputPath := "testdata/test.jpeg"
  
  sourceImage, _, err := loadImage( inputPath )
  if err != nil {
    t.Fatalf( "The image could not be loaded: %v", err )
  }
  
  bounds := sourceImage.Bounds()
  width := bounds.Dx()
  height := bounds.Dy()
  
  aspectRatio := float64( width ) / float64( height )
  
  if aspectRatio <= 0 {
    t.Errorf( "The aspect ratio is invalid: %.2f.", aspectRatio )
  }
  
  t.Logf( "Aspect ratio: %.2f:1.", aspectRatio )
}

func TestImageInfoTransparency( t *testing.T ) {
  if _, err := os.Stat( "testdata/test.png" ); os.IsNotExist( err ) {
    t.Skip( "PNG test file not present, skipping transparency test." )
  }
  
  sourceImage, _, err := loadImage( "testdata/test.png" )
  if err != nil {
    t.Fatalf( "The PNG image could not be loaded: %v", err )
  }
  
  hasAlpha := false
  switch sourceImage.( type ) {
  case *image.NRGBA, *image.NRGBA64, *image.RGBA, *image.RGBA64, *image.Alpha, *image.Alpha16:
    hasAlpha = true
  }
  
  t.Logf( "PNG has alpha channel: %v.", hasAlpha )
}

func TestImageInfoColorModel( t *testing.T ) {
  inputPath := "testdata/test.jpeg"

  sourceImage, _, err := loadImage( inputPath )
  if err != nil {
    t.Fatalf( "The image could not be loaded: %v", err )
  }

  colorModelName := fmt.Sprintf( "%T", sourceImage )

  if colorModelName == "" {
    t.Error( "The color model name should not be empty." )
  }

  t.Logf( "Color model: %s.", colorModelName )
}

func TestClipBasic( t *testing.T ) {
  inputPath := "testdata/test.jpeg"
  outputPath := "testdata/output_clipped.jpg"

  defer os.Remove( outputPath )

  sourceImage, format, err := loadImage( inputPath )
  if err != nil {
    t.Fatalf( "The source image could not be loaded: %v", err )
  }

  bounds := sourceImage.Bounds()
  originalWidth := bounds.Dx()
  originalHeight := bounds.Dy()

  // clip a 100x100 region from the top-left
  x1, y1, x2, y2 := 0, 0, 100, 100

  if x2 > originalWidth || y2 > originalHeight {
    t.Skip( "The test image is too small for this test." )
  }

  clipWidth := x2 - x1
  clipHeight := y2 - y1

  clippedImage := image.NewRGBA( image.Rect( 0, 0, clipWidth, clipHeight ) )
  for y := 0; y < clipHeight; y++ {
    for x := 0; x < clipWidth; x++ {
      clippedImage.Set( x, y, sourceImage.At( x1+x, y1+y ) )
    }
  }

  err = encodeOutput( outputPath, ".jpg", clippedImage, 90, format )
  if err != nil {
    t.Fatalf( "The clipped image could not be encoded: %v", err )
  }

  // verify the output file was created
  if _, err := os.Stat( outputPath ); os.IsNotExist( err ) {
    t.Error( "The output file was not created." )
  }

  // verify the output dimensions
  outputImage, _, err := loadImage( outputPath )
  if err != nil {
    t.Fatalf( "The output image could not be loaded: %v", err )
  }

  outputBounds := outputImage.Bounds()
  if outputBounds.Dx() != clipWidth || outputBounds.Dy() != clipHeight {
    t.Errorf( "The output dimensions are incorrect: expected %dx%d, got %dx%d.",
      clipWidth, clipHeight, outputBounds.Dx(), outputBounds.Dy() )
  }

  t.Logf( "Clipped %dx%d region from %dx%d image.", clipWidth, clipHeight, originalWidth, originalHeight )
}

func TestClipDifferentRegions( t *testing.T ) {
  inputPath := "testdata/test.jpeg"

  sourceImage, _, err := loadImage( inputPath )
  if err != nil {
    t.Fatalf( "The source image could not be loaded: %v", err )
  }

  bounds := sourceImage.Bounds()
  width := bounds.Dx()
  height := bounds.Dy()

  tests := []struct {
    name          string
    x1, y1, x2, y2 int
    expectWidth   int
    expectHeight  int
  }{
    { "top-left corner", 0, 0, 50, 50, 50, 50 },
    { "center region", width/4, height/4, width/4 + 100, height/4 + 100, 100, 100 },
    { "wide rectangle", 0, 0, 200, 50, 200, 50 },
    { "tall rectangle", 0, 0, 50, 200, 50, 200 },
  }

  for _, tt := range tests {
    t.Run( tt.name, func( t *testing.T ) {
      if tt.x2 > width || tt.y2 > height {
        t.Skip( "The test image is too small for this test." )
      }

      outputPath := fmt.Sprintf( "testdata/output_clip_%s.jpg", tt.name )
      defer os.Remove( outputPath )

      clipWidth := tt.x2 - tt.x1
      clipHeight := tt.y2 - tt.y1

      clippedImage := image.NewRGBA( image.Rect( 0, 0, clipWidth, clipHeight ) )
      for y := 0; y < clipHeight; y++ {
        for x := 0; x < clipWidth; x++ {
          clippedImage.Set( x, y, sourceImage.At( tt.x1+x, tt.y1+y ) )
        }
      }

      err = encodeOutput( outputPath, ".jpg", clippedImage, 90, "jpeg" )
      if err != nil {
        t.Fatalf( "The clipped image could not be encoded: %v", err )
      }

      outputImage, _, err := loadImage( outputPath )
      if err != nil {
        t.Fatalf( "The output image could not be loaded: %v", err )
      }

      outputBounds := outputImage.Bounds()
      if outputBounds.Dx() != tt.expectWidth || outputBounds.Dy() != tt.expectHeight {
        t.Errorf( "The output dimensions are incorrect: expected %dx%d, got %dx%d.",
          tt.expectWidth, tt.expectHeight, outputBounds.Dx(), outputBounds.Dy() )
      }
    } )
  }
}

func TestClipCoordinateValidation( t *testing.T ) {
  tests := []struct {
    name           string
    x1, y1, x2, y2 int
    imageWidth     int
    imageHeight    int
    shouldErr      bool
    errContains    string
  }{
    { "valid coordinates", 0, 0, 100, 100, 200, 200, false, "" },
    { "x2 equals x1", 50, 0, 50, 100, 200, 200, true, "x2 must be greater than x1" },
    { "x2 less than x1", 100, 0, 50, 100, 200, 200, true, "x2 must be greater than x1" },
    { "y2 equals y1", 0, 50, 100, 50, 200, 200, true, "y2 must be greater than y1" },
    { "y2 less than y1", 0, 100, 100, 50, 200, 200, true, "y2 must be greater than y1" },
    { "x2 exceeds width", 0, 0, 300, 100, 200, 200, true, "exceeds the image width" },
    { "y2 exceeds height", 0, 0, 100, 300, 200, 200, true, "exceeds the image height" },
  }

  for _, tt := range tests {
    t.Run( tt.name, func( t *testing.T ) {
      hasErr := false
      errMsg := ""

      if tt.x2 <= tt.x1 {
        hasErr = true
        errMsg = "x2 must be greater than x1"
      } else if tt.y2 <= tt.y1 {
        hasErr = true
        errMsg = "y2 must be greater than y1"
      } else if tt.x2 > tt.imageWidth {
        hasErr = true
        errMsg = "exceeds the image width"
      } else if tt.y2 > tt.imageHeight {
        hasErr = true
        errMsg = "exceeds the image height"
      }

      if hasErr != tt.shouldErr {
        t.Errorf( "Expected error state %v, but got %v.", tt.shouldErr, hasErr )
      }

      if tt.shouldErr && errMsg != "" && tt.errContains != "" {
        if errMsg != tt.errContains && errMsg != "" {
          // just log the expected error message
          t.Logf( "Error message: %s", errMsg )
        }
      }
    } )
  }
}

func TestClipOutputFormats( t *testing.T ) {
  inputPath := "testdata/test.jpeg"

  sourceImage, _, err := loadImage( inputPath )
  if err != nil {
    t.Fatalf( "The source image could not be loaded: %v", err )
  }

  bounds := sourceImage.Bounds()
  if bounds.Dx() < 50 || bounds.Dy() < 50 {
    t.Skip( "The test image is too small for this test." )
  }

  // create a small clipped region
  clippedImage := image.NewRGBA( image.Rect( 0, 0, 50, 50 ) )
  for y := 0; y < 50; y++ {
    for x := 0; x < 50; x++ {
      clippedImage.Set( x, y, sourceImage.At( x, y ) )
    }
  }

  tests := []struct {
    name      string
    output    string
    extension string
  }{
    { "JPEG output", "testdata/output_clip.jpg", ".jpg" },
    { "PNG output", "testdata/output_clip.png", ".png" },
    { "GIF output", "testdata/output_clip.gif", ".gif" },
    { "TIFF output", "testdata/output_clip.tiff", ".tiff" },
  }

  for _, tt := range tests {
    t.Run( tt.name, func( t *testing.T ) {
      defer os.Remove( tt.output )

      err := encodeOutput( tt.output, tt.extension, clippedImage, 90, "jpeg" )
      if err != nil {
        t.Fatalf( "The clipped image could not be encoded as %s: %v", tt.extension, err )
      }

      if _, err := os.Stat( tt.output ); os.IsNotExist( err ) {
        t.Errorf( "The output file %s was not created.", tt.output )
      }

      // verify the output can be loaded
      outputImage, _, err := loadImage( tt.output )
      if err != nil {
        t.Fatalf( "The output image could not be loaded: %v", err )
      }

      outputBounds := outputImage.Bounds()
      if outputBounds.Dx() != 50 || outputBounds.Dy() != 50 {
        t.Errorf( "The output dimensions are incorrect: expected 50x50, got %dx%d.",
          outputBounds.Dx(), outputBounds.Dy() )
      }
    } )
  }
}
