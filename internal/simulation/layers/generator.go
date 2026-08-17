package layers

import (
	"math"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/rng"
	"github.com/julianstephens/go-utils/logger"
)

const (
	baseFertility          = 0.35
	moistureWeight         = 0.45
	elevationPenaltyWeight = 0.40
	variationAmplitude     = 0.12
)

type Generator struct {
	width, height int
	random        *rng.RNG
}

type LayerArtifacts struct {
	Elevation *simtypes.Layer
	Moisture  *simtypes.Layer
	Fertility *simtypes.Layer
	Terrain   *simtypes.Layer
}

func NewGenerator(width, height int, randomSource *rng.RNG) *Generator {
	return &Generator{
		width:  width,
		height: height,
		random: randomSource,
	}
}

func (g *Generator) GenerateLayers(worldSeed [32]byte, _ *simtypes.TerrainMap) (LayerArtifacts, error) {
	logger.WithFields(map[string]interface{}{"width": g.width, "height": g.height}).Debug("generating elevation layer")
	elevationLayer := generateElevationLayer(g, worldSeed)

	logger.WithFields(map[string]interface{}{"width": g.width, "height": g.height}).Debug("generating moisture layer")
	moistureLayer := generateMoistureLayer(g, worldSeed)

	logger.WithFields(map[string]interface{}{"width": g.width, "height": g.height}).Debug("generating fertility layer")
	fertilityLayer := generateFertilityLayer(g, worldSeed, elevationLayer, moistureLayer)

	// Terrain classification is owned by TerrainClassifier. The generator only
	// produces continuous scalar layers.
	terrainLayer := simtypes.NewLayer(g.width, g.height)

	return LayerArtifacts{
		Elevation: elevationLayer,
		Moisture:  moistureLayer,
		Fertility: fertilityLayer,
		Terrain:   terrainLayer,
	}, nil
}

func generateElevationLayer(g *Generator, worldSeed [32]byte) *simtypes.Layer {
	elevationSeed := rng.DeriveSeed(worldSeed, "elevation")
	elevationGenerator := NewDSGenerator(g.width, rng.NewRNG(elevationSeed))
	return normalizeDSOutput(elevationGenerator.Generate())
}

func generateMoistureLayer(g *Generator, worldSeed [32]byte) *simtypes.Layer {
	moistureSeed := rng.DeriveSeed(worldSeed, "moisture")
	moistureGenerator := NewDSGenerator(g.width, rng.NewRNG(moistureSeed))
	return normalizeDSOutput(moistureGenerator.Generate())
}

func generateFertilityLayer(
	g *Generator,
	worldSeed [32]byte,
	elevationLayer *simtypes.Layer,
	moistureLayer *simtypes.Layer,
) *simtypes.Layer {
	fertilityLayer := simtypes.NewLayer(g.width, g.height)
	noiseSeed := rng.DeriveSeed(worldSeed, "fertility")
	noiseRNG := rng.NewRNG(noiseSeed)

	for x := 0; x < g.width; x++ {
		for y := 0; y < g.height; y++ {
			elevation := elevationLayer.Get(x, y)
			moisture := moistureLayer.Get(x, y)

			extremeElevation := 2.0 * math.Abs(elevation-0.5)
			penalty := extremeElevation * extremeElevation * elevationPenaltyWeight
			noise := (noiseRNG.Elevation(nil) - 0.5) * 2.0
			variation := noise * variationAmplitude

			fertility := baseFertility + (moisture * moistureWeight) - penalty + variation
			fertilityLayer.Set(x, y, clamp(fertility))
		}
	}

	return fertilityLayer
}

func normalizeLayer(layer *simtypes.Layer) *simtypes.Layer {
	minValue, maxValue := minMax(layer)
	if maxValue == minValue {
		return layer
	}

	for i, value := range layer.Values {
		layer.Values[i] = clamp((value - minValue) / (maxValue - minValue))
	}
	return layer
}

func minMax(layer *simtypes.Layer) (float64, float64) {
	minValue := layer.Get(0, 0)
	maxValue := minValue

	for _, value := range layer.Values {
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}

	return minValue, maxValue
}

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
