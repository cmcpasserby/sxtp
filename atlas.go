package sxtp

import (
	"fmt"
	"image"
	"math"
)

const (
	deg2rad = math.Pi / 180
)

type Angle float64

func (a Angle) Radians() float64 {
	return float64(a) * deg2rad
}

func (a Angle) Degrees() float64 {
	return float64(a)
}

type Filter struct {
	X, Y string
}

func (f *Filter) String() string {
	return fmt.Sprintf("%s,%s", f.X, f.Y)
}

type Bounds struct {
	Position image.Point // Position in Image
	Size     image.Point // Packed Size
}

func (b *Bounds) String() string {
	return fmt.Sprintf("Position: %v | Size: %v", b.Position, b.Size)
}

type Offsets struct {
	Offset       image.Point // Offset (Left, Bottom)
	OriginalSize image.Point // Original Size
}

func (o *Offsets) String() string {
	return fmt.Sprintf("Offset: %v | OriginalSize: %v", o.Offset, o.OriginalSize)
}

type Atlas struct {
	Name    string      `atlas:"name"`
	Size    image.Point `atlas:"size"`
	Format  string      `atlas:"format"`
	Filter  Filter      `atlas:"filter"`
	Repeat  string      `atlas:"repeat"`
	Pma     bool        `atlas:"pma"`
	Sprites []Sprite    `atlas:"sprites"`
}

func (a *Atlas) String() string {
	return a.Name
}

type Sprite struct {
	Name    string  `atlas:"name"`
	Index   int     `atlas:"index"`
	Bounds  Bounds  `atlas:"bounds"`
	Offsets Offsets `atlas:"offsets"`
	Rotate  Angle   `atlas:"rotate"`
	// Split   image.Rectangle `atlas:"split"`
	// Pad     image.Rectangle `atlas:"pad"`
}

func (a *Sprite) String() string {
	return a.Name
}
