package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/julianstephens/aeon/internal/simtypes"
)

const scale = 8

func main() {
	terrainMap := generateTerrainMap()

	img := renderTerrainMap(terrainMap, scale)

	file, err := os.Create("terrain.png")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	if err := png.Encode(file, img); err != nil {
		log.Fatal(err)
	}
}

func renderTerrainMap(tm simtypes.TerrainMap, scale int) image.Image {
	width := tm.Width * scale
	height := tm.Height * scale

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			cell := tm.GetCell(x, y)
			if cell == nil {
				continue
			}

			c := terrainColor(cell.Terrain)

			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					img.Set(
						x*scale+dx,
						y*scale+dy,
						c,
					)
				}
			}
		}
	}

	return img
}

func terrainColor(terrain simtypes.TerrainType) color.Color {
	switch terrain {
	case simtypes.TerrainTypePlains:
		return color.RGBA{R: 190, G: 170, B: 100, A: 255}
	case simtypes.TerrainTypeForest:
		return color.RGBA{R: 50, G: 120, B: 60, A: 255}
	case simtypes.TerrainTypeMountain:
		return color.RGBA{R: 100, G: 100, B: 100, A: 255}
	case simtypes.TerrainTypeWater:
		return color.RGBA{R: 50, G: 100, B: 180, A: 255}
	default:
		return color.Black
	}
}

func generateTerrainMap() simtypes.TerrainMap {
	// Replace this with your actual generator.
	tm := simtypes.NewTerrainMap(
		simtypes.DefaultMapWidth,
		simtypes.DefaultMapHeight,
	)

	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			cell := tm.GetCell(x, y)
			cell.Location = simtypes.Position{X: x, Y: y}

			// Temporary test pattern.
			switch {
			case y < 15:
				cell.Terrain = simtypes.TerrainTypeWater
			case x > 45 && y > 20:
				cell.Terrain = simtypes.TerrainTypeMountain
			case x > 20 && x < 40 && y > 20 && y < 45:
				cell.Terrain = simtypes.TerrainTypeForest
			default:
				cell.Terrain = simtypes.TerrainTypePlains
			}
		}
	}

	return tm
}
