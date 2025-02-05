package sxtp

import (
	"fmt"
	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
	"golang.org/x/sync/errgroup"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type FileFormat string

const (
	FormatPNG FileFormat = "png"
	FormatJPG FileFormat = "jpg"
)

var (
	supportedFormats = [...]FileFormat{FormatPNG, FormatJPG}
)

// PackMasks packs a set of secondary textures using the atlas layouts passed in
func PackMasks(atlasPages []Atlas, format FileFormat, masksPath, outputPath, suffix string, hasAlpha bool, l *log.Logger) error {
	maskImages, err := getImageNames(masksPath)
	if err != nil {
		return err
	}

	var wg errgroup.Group

	for _, atlas := range atlasPages {
		atlas := &atlas

		wg.Go(func() error {
			noExtName := strings.TrimSuffix(atlas.Name, filepath.Ext(atlas.Name))
			maskFileName := fmt.Sprintf("%s_%s.%s", noExtName, suffix, format)
			outFileName := filepath.Join(outputPath, maskFileName)
			return packPage(atlas, maskImages, format, outFileName, hasAlpha, l)
		})
	}

	return wg.Wait()
}

func packPage(atlas *Atlas, maskImages map[string]string, format FileFormat, outputPath string, hasAlpha bool, l *log.Logger) error {
	outImage := image.NewNRGBA(image.Rectangle{Max: atlas.Size})
	packedCount := 0

	for _, sprite := range atlas.Sprites {
		maskImagePath, ok := maskImages[sprite.Name]
		if !ok {
			continue
		}

		maskImage, err := loadImage(maskImagePath, l)
		if err != nil {
			return err
		}

		// TODO combine this with the crop and offset and try to do it with mostly 1 transform also apply scale
		if sprite.Rotate != 0 {
			size := maskImage.Bounds().Size()
			rotatedMask := image.NewNRGBA(image.Rect(0, 0, size.Y, size.X))

			draw.CatmullRom.Transform(rotatedMask, makeRotation(-sprite.Rotate.Radians(), rotatedMask.Bounds().Size()), maskImage, maskImage.Bounds(), draw.Over, nil)
			maskImage = rotatedMask
		}

		maskImage = cropAndOffset(maskImage, sprite)

		spriteRect := image.Rectangle{Min: sprite.Bounds.Position, Max: sprite.Bounds.Position.Add(maskImage.Bounds().Size())}
		draw.Draw(outImage, spriteRect, maskImage, image.Point{}, draw.Over) // TODO might need to use DrawMask, and mask with color image alpha

		if !hasAlpha {
			stripAlpha(outImage)
		}

		packedCount++
	}

	if err := saveImage(outputPath, outImage, format, l); err != nil {
		return err
	}

	l.Printf("Packed %d sprites to %s\n", packedCount, outputPath)
	return nil
}

func rotateCoord(pt image.Point) image.Point {
	return image.Pt(pt.Y, pt.X)
}

func cropAndOffset(img image.Image, sprite Sprite) image.Image {
	if sprite.Offsets.Offset.Eq(image.Point{}) && sprite.Bounds.Size.Eq(sprite.Offsets.OriginalSize) {
		return img
	}

	size := sprite.Bounds.Size
	orig := sprite.Offsets.OriginalSize
	offset := sprite.Offsets.Offset

	var newRect image.Rectangle
	var sp image.Point

	if sprite.Rotate == 90 {
		newRect = image.Rectangle{Max: rotateCoord(size)}
		sp = rotateCoord(orig.Sub(size).Sub(offset))
	} else {
		newRect = image.Rectangle{Max: size}
		sp = image.Pt(offset.X, orig.Y-size.Y-offset.Y)
	}

	cropped := image.NewNRGBA(newRect)
	draw.Draw(cropped, cropped.Rect, img, sp, draw.Over)

	return cropped
}

func stripAlpha(img *image.NRGBA) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := img.NRGBAAt(x, y)
			c.A = 0xff
			img.SetNRGBA(x, y, c)
		}
	}
}

func getImageNames(path string) (map[string]string, error) {
	images := make([]string, 0)

	err := filepath.Walk(path, func(p string, info fs.FileInfo, err error) error {
		ext := filepath.Ext(p)

		for _, format := range supportedFormats {
			if fmt.Sprintf(".%s", format) == ext {
				images = append(images, p)
				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	mapping := make(map[string]string, len(images))

	for _, img := range images {
		name, err := filepath.Rel(path, img)
		if err != nil {
			return nil, err
		}
		name = strings.TrimSuffix(name, filepath.Ext(name))
		name = strings.ReplaceAll(name, "\\", "/")
		mapping[name] = img
	}

	return mapping, nil
}

func loadImage(path string, l *log.Logger) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer closeWithLoggedError(f, l)

	img, _, err := image.Decode(f)
	return img, err
}

func saveImage(path string, img image.Image, format FileFormat, l *log.Logger) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer closeWithLoggedError(f, l)

	switch format {
	case FormatPNG:
		encoder := png.Encoder{CompressionLevel: png.NoCompression}
		return encoder.Encode(f, img)

	case FormatJPG:
		options := jpeg.Options{Quality: 100}
		return jpeg.Encode(f, img, &options)

	default:
		return fmt.Errorf("invalid file format, got: %q expected: png or jpg", format)
	}
}

func closeWithLoggedError(c io.Closer, l *log.Logger) {
	if err := c.Close(); err != nil {
		l.Println(err)
	}
}

func makeRotation(r float64, size image.Point) f64.Aff3 {
	cosD, sinD := math.Cos(r), math.Sin(r)

	return f64.Aff3{
		cosD, -sinD, 0,
		sinD, cosD, float64(size.Y),
	}
}
