package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/layers"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

type LayerType string

const (
	LayerTypeTerrain      LayerType = "terrain"
	LayerTypeElevation    LayerType = "elevation"
	LayerTypeMoisture     LayerType = "moisture"
	LayerTypeFertility    LayerType = "fertility"
	LayerTypeFoodCapacity LayerType = "food-capacity"
)

const defaultScale = 8

func main() {
	seed := flag.Uint64("seed", 42, "Terrain generation seed")
	layerName := flag.String("layer", string(LayerTypeElevation), "Layer to render")
	scale := flag.Int("scale", defaultScale, "Scale factor for each map cell")
	output := flag.String("output", "terrain.png", "Output PNG path")

	flag.Parse()

	if *scale < 1 {
		log.Fatal("scale must be at least 1")
	}

	tm, elevation := generateTerrainMap(*seed)

	img, err := renderLayer(tm, elevation, LayerType(*layerName), *scale)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := png.Encode(file, img); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("wrote %s\n", *output)
}

func generateTerrainMap(seed uint64) (simtypes.TerrainMap, *simtypes.ElevationMap) {
	seedBytes := sha256.Sum256([]byte(fmt.Sprintf("%d", seed)))
	random := rng.NewRNG(seedBytes)

	tm := simtypes.NewTerrainMap(
		simtypes.DefaultMapWidth,
		simtypes.DefaultMapHeight,
	)

	elevationGenerator := layers.NewDSGenerator(max(tm.Width, tm.Height), random)
	elevation := elevationGenerator.Generate()
	tm.ApplyElevation(*elevation)

	return tm, elevation
}

func renderLayer(
	tm simtypes.TerrainMap,
	elevation *simtypes.ElevationMap,
	layer LayerType,
	scale int,
) (image.Image, error) {
	switch layer {
	case LayerTypeTerrain:
		return renderTerrainMap(tm, scale), nil
	case LayerTypeElevation:
		return renderElevationMap(*elevation, tm.Width, tm.Height, scale), nil
	case LayerTypeMoisture:
		return renderScalarLayer(tm.Width, tm.Height, scale, func(x, y int) float64 {
			cell := tm.GetCell(x, y)
			return cell.Moisture
		}), nil
	case LayerTypeFertility:
		return renderScalarLayer(tm.Width, tm.Height, scale, func(x, y int) float64 {
			cell := tm.GetCell(x, y)
			return cell.Fertility
		}), nil
	case LayerTypeFoodCapacity:
		return renderScalarLayer(tm.Width, tm.Height, scale, func(x, y int) float64 {
			cell := tm.GetCell(x, y)
			return cell.FoodCapacity
		}), nil
	default:
		return nil, fmt.Errorf("unknown layer %q; valid layers: %s, %s, %s, %s, %s", layer, LayerTypeTerrain, LayerTypeElevation, LayerTypeMoisture, LayerTypeFertility, LayerTypeFoodCapacity)
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

			fillCell(img, x, y, scale, terrainColor(cell.Terrain))
		}
	}

	return img
}

func renderElevationMap(em simtypes.ElevationMap, width, height, scale int) image.Image {
	img := image.NewGray(image.Rect(0, 0, width*scale, height*scale))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x >= em.Width || y >= em.Height {
				continue
			}

			value := normalize(em.Get(x, y))
			fillCell(img, x, y, scale, color.Gray{Y: uint8(value * 255)})
		}
	}

	return img
}

func renderScalarLayer(width, height, scale int, valueAt func(x, y int) float64) image.Image {
	img := image.NewGray(image.Rect(0, 0, width*scale, height*scale))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := normalize(valueAt(x, y))
			fillCell(img, x, y, scale, color.Gray{Y: uint8(value * 255)})
		}
	}

	return img
}

func fillCell(img interface{ Set(x, y int, c color.Color) }, x, y, scale int, c color.Color) {
	for dy := 0; dy < scale; dy++ {
		for dx := 0; dx < scale; dx++ {
			img.Set(x*scale+dx, y*scale+dy, c)
		}
	}
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

func normalize(value float64) float64 {
	switch {
	case value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
