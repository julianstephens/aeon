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
	"strconv"
	"strings"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/layers"
	"github.com/julianstephens/aeon/internal/simulation/rng"
	"github.com/julianstephens/go-utils/cliutil"
	"github.com/julianstephens/go-utils/logger"
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
	args := cliutil.ParseArgs(os.Args[1:])
	configureLogLevel(args)

	seedStr := args.GetFlagWithDefault("seed", "42")
	seed, err := strconv.ParseUint(seedStr, 10, 64)
	if err != nil {
		logger.Fatalf("invalid seed: %v", err)
	}

	layerName := args.GetFlagWithDefault("layer", string(LayerTypeElevation))

	scaleStr := args.GetFlagWithDefault("scale", strconv.Itoa(defaultScale))
	scale, err := strconv.Atoi(scaleStr)
	if err != nil {
		logger.Fatalf("invalid scale: %v", err)
	}

	output := args.GetFlagWithDefault("output", "terrain.png")
	outputPath, err := sanitizeOutputPath(output)
	if err != nil {
		logger.Fatalf("invalid output path: %v", err)
	}

	if scale < 1 {
		logger.Fatal("scale must be at least 1")
	}

	logger.WithFields(map[string]interface{}{
		"seed":   seed,
		"layer":  layerName,
		"scale":  scale,
		"output": outputPath,
	}).Info("generating map visualization")

	tm, elevation := generateTerrainMap(seed)

	img, err := renderLayer(tm, elevation, LayerType(layerName), scale)
	if err != nil {
		logger.Fatal(err)
	}

	outputDir := filepath.Dir(outputPath)
	if outputDir != "." {
		// #nosec G703 -- outputDir is derived from sanitizeOutputPath-validated relative outputPath.
		if err := os.MkdirAll(outputDir, 0o750); err != nil {
			logger.Fatalf("failed to create output directory: %v", err)
		}
	}

	// #nosec G304 G703 -- outputPath is sanitized by sanitizeOutputPath and restricted to a relative .png path.
	file, err := os.Create(outputPath)
	if err != nil {
		logger.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Errorf("failed to close output file: %v", err)
		}
	}()

	if err := png.Encode(file, img); err != nil {
		logger.Fatal(err)
	}

	logger.Infof("wrote %s", outputPath)
}

func sanitizeOutputPath(raw string) (string, error) {
	cleaned := filepath.Clean(strings.TrimSpace(raw))
	if cleaned == "" || cleaned == "." {
		return "", fmt.Errorf("output path is empty")
	}
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("absolute output paths are not allowed")
	}

	ext := strings.ToLower(filepath.Ext(cleaned))
	if ext == "" {
		cleaned += ".png"
		ext = ".png"
	}
	if ext != ".png" {
		return "", fmt.Errorf("output file must use .png extension")
	}

	return cleaned, nil
}

func configureLogLevel(args *cliutil.Args) {
	level := args.GetFlag("log-level")
	if level == "" {
		level = os.Getenv("AEON_LOG_LEVEL")
	}
	if level == "" {
		level = "info"
	}

	if err := logger.SetLogLevel(level); err != nil {
		logger.Warnf("invalid log level %q, falling back to info", level)
		_ = logger.SetLogLevel("info")
	}
}

func generateTerrainMap(seed uint64) (simtypes.TerrainMap, *simtypes.ElevationMap) {
	seedBytes := sha256.Sum256(fmt.Appendf(nil, "%d", seed))
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
		return nil, fmt.Errorf(
			"unknown layer %q; valid layers: %s, %s, %s, %s, %s",
			layer,
			LayerTypeTerrain,
			LayerTypeElevation,
			LayerTypeMoisture,
			LayerTypeFertility,
			LayerTypeFoodCapacity,
		)
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
	minValue, maxValue := math.Inf(1), math.Inf(-1)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x >= em.Width || y >= em.Height {
				continue
			}

			v := em.Get(x, y)
			if v < minValue {
				minValue = v
			}
			if v > maxValue {
				maxValue = v
			}
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x >= em.Width || y >= em.Height {
				continue
			}

			value := normalizeToRange(em.Get(x, y), minValue, maxValue)
			fillCell(img, x, y, scale, color.Gray{Y: uint8(value * 255)})
		}
	}

	return img
}

func renderScalarLayer(width, height, scale int, valueAt func(x, y int) float64) image.Image {
	img := image.NewGray(image.Rect(0, 0, width*scale, height*scale))
	minValue, maxValue := math.Inf(1), math.Inf(-1)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			v := valueAt(x, y)
			if v < minValue {
				minValue = v
			}
			if v > maxValue {
				maxValue = v
			}
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

	switch {
	case normalized < 0:
		return 0
	case normalized > 1:
		return 1
	default:
		return normalized
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
