//go:build nintendo3ds

package main

import (
	"image"
	"image/color"
	"math"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/scene"
)

// mapSize is the map's side in pixels, and mapExtent the half side it covers on the ground: the
// island's terrain reaches 120 units from its centre.
const (
	mapSize   = 96
	mapExtent = 120
)

// heightColours colour the map by height, as a chart would: sea, shore, grass, forest, rock,
// snow. Heights in the island's units; colours in the screen's gamma.
var heightColours = []struct {
	height float64
	colour color.NRGBA
}{
	{-8, color.NRGBA{0x17, 0x4a, 0x63, 0xff}},
	{0, color.NRGBA{0x2f, 0x7f, 0x96, 0xff}},
	{0.5, color.NRGBA{0xd8, 0xc8, 0x98, 0xff}},
	{3, color.NRGBA{0x7d, 0xa8, 0x5c, 0xff}},
	{12, color.NRGBA{0x3f, 0x76, 0x46, 0xff}},
	{22, color.NRGBA{0x8a, 0x8a, 0x80, 0xff}},
	{31.5, color.NRGBA{0xf0, 0xf3, 0xf6, 0xff}},
}

// islandMap draws the island from above, coloured by height and shaded as if lit from the
// north-west, with a dot on each place the camera can fly to.
func islandMap(world *scene.Scene) *draw2d.Image {
	picture := image.NewNRGBA(image.Rect(0, 0, mapSize, mapSize))
	height := func(column, row int) float64 {
		x, z := mapGround(column), mapGround(row)
		ground, found := world.GroundHeight(float32(x), float32(z))
		if !found {
			return -8
		}
		return float64(ground)
	}
	for row := range mapSize {
		for column := range mapSize {
			here := height(column, row)
			slope := height(column+1, row+1) - height(column-1, row-1)
			shade := 1 - math.Max(-0.35, math.Min(0.35, slope*0.06))
			if here <= 0 {
				shade = 1 // the sea is flat
			}
			picture.SetNRGBA(column, row, shaded(colourForHeight(here), shade))
		}
	}
	for _, destination := range places[1:] {
		dot(picture, mapPixel(float64(destination.target[0])), mapPixel(float64(destination.target[2])))
	}
	return draw2d.NewImage(picture)
}

// mapGround is where on the ground a column or row of the map is.
func mapGround(pixel int) float64 {
	return (float64(pixel)+0.5)/mapSize*2*mapExtent - mapExtent
}

// mapPixel is the column or row of the map over a point of the ground.
func mapPixel(ground float64) int {
	return int((ground + mapExtent) / (2 * mapExtent) * mapSize)
}

func colourForHeight(height float64) color.NRGBA {
	if height <= heightColours[0].height {
		return heightColours[0].colour
	}
	for index := 1; index < len(heightColours); index++ {
		upper := heightColours[index]
		if height > upper.height {
			continue
		}
		lower := heightColours[index-1]
		share := (height - lower.height) / (upper.height - lower.height)
		return mix(lower.colour, upper.colour, share)
	}
	return heightColours[len(heightColours)-1].colour
}

func mix(from, to color.NRGBA, share float64) color.NRGBA {
	channel := func(first, second uint8) uint8 {
		return uint8(float64(first) + (float64(second)-float64(first))*share + 0.5)
	}
	return color.NRGBA{channel(from.R, to.R), channel(from.G, to.G), channel(from.B, to.B), 0xff}
}

func shaded(colour color.NRGBA, shade float64) color.NRGBA {
	channel := func(value uint8) uint8 { return uint8(math.Min(255, float64(value)*shade+0.5)) }
	return color.NRGBA{channel(colour.R), channel(colour.G), channel(colour.B), 0xff}
}

// dot marks a place: white, ringed in dark, three pixels across.
func dot(picture *image.NRGBA, column, row int) {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			distance := dx*dx + dy*dy
			switch {
			case distance <= 2:
				picture.SetNRGBA(column+dx, row+dy, color.NRGBA{0xff, 0xff, 0xff, 0xff})
			case distance <= 5:
				picture.SetNRGBA(column+dx, row+dy, color.NRGBA{0x0b, 0x18, 0x20, 0xff})
			}
		}
	}
}
