package main

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
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

type CLI struct {
	Seed      uint64    `help:"Terrain generation seed." default:"42"`
	Layer     LayerType `help:"Layer to render." enum:"terrain,elevation,moisture,fertility,food-capacity" default:"elevation"`
	Scale     int       `help:"Scale factor for each map cell." default:"8"`
	Output    string    `help:"Output PNG path." default:"terrain.png"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("mapviz"),
		kong.Description("Generate PNG visualizations of Aeon terrain layers."),
	)
	ctx.FatalIfError2()

	if cli.Scale < 1 {
		ctx.Fatalf("scale must be at least 1")
	}

	outputPath, err := sanitizeOutputPath(cli.Output)
	if err != nil {
		ctx.Fatalf("invalid output path: %v", err)
	}

	tm, elevation, err := generateTerrainMap(cli.Seed)
	if err != nil {
		ctx.Fatalf("failed to generate terrain map: %v", err)
	}

	img, err := renderLayer(*tm, elevation, cli.Layer, cli.Scale)
	if err != nil {
		ctx.Fatalf("failed to render layer: %v", err)
	}

	outputDir := filepath.Dir(outputPath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0o750); err != nil {
			ctx.Fatalf("failed to create output directory: %v", err)
		}
	}

	file, err := os.Create(outputPath) // #nosec G304 -- outputPath is restricted to relative .png paths.
	if err != nil {
		ctx.Fatalf("failed to create output file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := png.Encode(file, img); err != nil {
		ctx.Fatalf("failed to encode PNG: %v", err)
	}

	fmt.Printf("wrote %s\n", outputPath)
}

func sanitizeOutputPath(raw string) (string, error) {
	cleaned := filepath.Clean(raw)
	if cleaned == "" || cleaned == "." {
		return "", fmt.Errorf("output path is empty")
	}
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("absolute output paths are not allowed")
	}
	if ext := filepath.Ext(cleaned); ext != ".png" {
		return "", fmt.Errorf("output file must use .png extension")
	}
	return cleaned, nil
}

func generateTerrainMap(seed uint64) (*simtypes.TerrainMap, *simtypes.Layer, error) {
	seedBytes := sha256.Sum256(fmt.Appendf(nil, "%d", seed))
	random := rng.NewRNG(seedBytes)

	pipeline := layers.NewPipeline(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight, random)
	tm, err := pipeline.Run(seedBytes)
	if err != nil {
		return nil, nil, err
	}

	generator := layers.NewGenerator(tm.Width, tm.Height, random)
	artifacts, err := generator.GenerateLayers(seedBytes, tm)
	if err != nil {
		return nil, nil, err
	}

	tm.ApplyElevation(artifacts.Elevation)
	tm.ApplyMoisture(artifacts.Moisture)
	tm.ApplyFertility(artifacts.Fertility)
	tm.ApplyTerrain(artifacts.Terrain)

	return tm, artifacts.Elevation, nil
}

func renderLayer(
	tm simtypes.TerrainMap,
	elevation *simtypes.Layer,
	layer LayerType,
	scale int,
) (image.Image, error) {
	switch layer {
	case LayerTypeTerrain:
		return renderTerrainMap(tm, scale), nil
	case LayerTypeElevation:
		return renderScalarLayer(tm.Width, tm.Height, scale, func(x, y int) float64 {
			return elevation.Get(x, y)
		}), nil
	case LayerTypeMoisture:
		return renderScalarLayer(tm.Width, tm.Height, scale, func(x, y int) float64 {
			return tm.GetCell(x, y).Moisture
		}), nil
	case LayerTypeFertility:
		return renderScalarLayer(tm.Width, tm.Height, scale, func(x, y int) float64 {
			return tm.GetCell(x, y).Fertility
		}), nil
	case LayerTypeFoodCapacity:
		return renderScalarLayer(tm.Width, tm.Height, scale, func(x, y int) float64 {
			return tm.GetCell(x, y).FoodCapacity
		}), nil
	default:
		return nil, fmt.Errorf("unsupported layer %q", layer)
	}
}

func renderTerrainMap(tm simtypes.TerrainMap, scale int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, tm.Width*scale, tm.Height*scale))

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

func renderScalarLayer(width, height, scale int, valueAt func(x, y int) float64) image.Image {
	img := image.NewGray(image.Rect(0, 0, width*scale, height*scale))
	minValue, maxValue := math.Inf(1), math.Inf(-1)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := valueAt(x, y)
			minValue = minFloat(minValue, value)
			maxValue = maxFloat(maxValue, value)
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := normalizeToRange(valueAt(x, y), minValue, maxValue)
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

func normalizeToRange(value, minValue, maxValue float64) float64 {
	if maxValue <= minValue {
		return 0.5
	}

	normalized := (value - minValue) / (maxValue - minValue)
	if normalized < 0 {
		return 0
	}
	if normalized > 1 {
		return 1
	}
	return normalized
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
