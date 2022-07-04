package sxtp

import (
	"fmt"
	"github.com/disintegration/imaging"
	"golang.org/x/sync/errgroup"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
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

func PackMasks(atlasPages []Atlas, format FileFormat, masksPath, outputPath, suffix string, hasAlpha bool, l *log.Logger) error {
	maskImages, err := getImageNames(masksPath)
	if err != nil {
		return err
	}

	var wg errgroup.Group

	for i, atlas := range atlasPages {
		i, atlas := i, &atlas

		wg.Go(func() error {
			noExtName := strings.TrimSuffix(atlas.Name, filepath.Ext(atlas.Name))
			maskFileName := fmt.Sprintf("%s_%s_%02d.%s", noExtName, suffix, i, format)
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
			l.Printf("Skipped %s since packed images path does not contain it", sprite.Name)
			continue
		}

		maskImage, err := loadImage(maskImagePath, l)
		if err != nil {
			return err
		}

		maskImage = imaging.Rotate(maskImage, sprite.Rotate.Degrees(), color.NRGBA{})
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
	for _, ext := range supportedFormats {
		globbed, err := filepath.Glob(fmt.Sprintf("%s/*.%s", path, ext))
		if err != nil {
			return nil, err
		}
		images = append(images, globbed...)
	}

	mapping := make(map[string]string, len(images))

	for _, img := range images {
		name := strings.TrimSuffix(filepath.Base(img), filepath.Ext(img))
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
