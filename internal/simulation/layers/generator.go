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

var (
	BaselineProductivity = map[simtypes.TerrainType]float64{
		simtypes.TerrainTypeWater:    0.0,
		simtypes.TerrainTypePlains:   0.8,
		simtypes.TerrainTypeForest:   0.6,
		simtypes.TerrainTypeMountain: 0.05,
	}
)

type Generator struct {
	width, height int
	random        *rng.RNG
}

type LayerArtifacts struct {
	Elevation *simtypes.Layer
	Moisture  *simtypes.Layer
	Fertility *simtypes.Layer
}

func NewGenerator(width, height int, randomSource *rng.RNG) *Generator {
	return &Generator{
		width:  width,
		height: height,
		random: randomSource,
	}
}

func (g *Generator) GenerateIntrinsicLayers(worldSeed [32]byte) (LayerArtifacts, error) {
	logger.WithFields(map[string]any{"width": g.width, "height": g.height}).Debug("generating elevation layer")
	elevationLayer := generateElevationLayer(g, worldSeed)

	logger.WithFields(map[string]any{"width": g.width, "height": g.height}).Debug("generating moisture layer")
	moistureLayer := generateMoistureLayer(g, worldSeed)

	logger.WithFields(map[string]any{"width": g.width, "height": g.height}).Debug("generating fertility layer")
	fertilityLayer := generateFertilityLayer(g, worldSeed, elevationLayer, moistureLayer)

	return LayerArtifacts{
		Elevation: elevationLayer,
		Moisture:  moistureLayer,
		Fertility: fertilityLayer,
	}, nil
}

func (g *Generator) GenerateComputedLayers(
	tm *simtypes.TerrainMap,
	elevationLayer, moistureLayer, fertilityLayer *simtypes.Layer,
) (foodCapacityLayer *simtypes.Layer) {
	logger.WithFields(map[string]any{"width": tm.Width, "height": tm.Height}).Debug("generating food capacity layer")
	foodCapacityLayer = generateFoodCapacityLayer(tm, elevationLayer, moistureLayer, fertilityLayer)

	return foodCapacityLayer
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

func generateFoodCapacityLayer(
	tm *simtypes.TerrainMap,
	elevationLayer, moistureLayer, fertilityLayer *simtypes.Layer,
) *simtypes.Layer {
	foodCapacityLayer := simtypes.NewLayer(fertilityLayer.Width, fertilityLayer.Height)

	for x := 0; x < tm.Width; x++ {
		for y := 0; y < tm.Height; y++ {
			cell := tm.GetCell(x, y)
			if cell == nil {
				continue
			}

			fertilityFactor := 0.5 + 0.5*fertilityLayer.Get(x, y)
			moistureFactor := clamp(1 - math.Abs((moistureLayer.Get(x, y)-0.6)/0.6))
			elevationFactor := 1 - 0.25*elevationLayer.Get(x, y)

			capacity := BaselineProductivity[cell.Terrain] * fertilityFactor * moistureFactor * elevationFactor
			foodCapacityLayer.Set(x, y, capacity)
		}
	}
	return foodCapacityLayer
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
