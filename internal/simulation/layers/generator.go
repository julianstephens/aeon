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

func NewGenerator(width, height int, randomSource *rng.RNG) *Generator {
	return &Generator{
		width:  width,
		height: height,
		random: randomSource,
	}
}

func (g *Generator) GenerateTerrainMap(worldSeed [32]byte, tm *simtypes.TerrainMap) error {
	// generate elevation map using Diamond-Square algorithm
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("generating elevation map")
	elevationMap := generateElevationMap(g, worldSeed)
	logger.WithFields(map[string]interface{}{
		"width":  elevationMap.Width,
		"height": elevationMap.Height,
	}).Debug("elevation map generated")

	// generate moisture map using Diamond-Square algorithm
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("generating moisture map")
	moistureMap := generateMoistureMap(g, worldSeed)
	logger.WithFields(map[string]interface{}{
		"width":  moistureMap.Width,
		"height": moistureMap.Height,
	}).Debug("moisture map generated")

	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("generating fertility map")
	fertilityMap := generateFertilityMap(g, worldSeed, elevationMap, moistureMap)
	logger.WithFields(map[string]interface{}{
		"width":  fertilityMap.Width,
		"height": fertilityMap.Height,
	}).Debug("fertility map generated")

	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("applying elevation map to terrain map")
	tm.ApplyElevation(elevationMap)
	logger.Debug("elevation map applied to terrain map")
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("applying moisture map to terrain map")
	tm.ApplyMoisture(moistureMap)
	logger.Debug("moisture map applied to terrain map")
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("applying fertility map to terrain map")
	tm.ApplyFertility(fertilityMap)
	logger.Debug("fertility map applied to terrain map")
	tm.SetInitialized(true)
	return nil
}

func generateElevationMap(g *Generator, worldSeed [32]byte) *simtypes.LayerMap {
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("deriving elevation seed from world seed")
	elevationSeed := rng.DeriveSeed(worldSeed, "elevation")
	elevationGenerator := NewDSGenerator(g.width, rng.NewRNG(elevationSeed))
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("seed derived for elevation map generation")
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("performing diamond-square algorithm to generate elevation map")
	elevationMap := elevationGenerator.Generate()
	return elevationMap
}

func generateMoistureMap(g *Generator, worldSeed [32]byte) *simtypes.LayerMap {
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("deriving moisture seed from world seed")
	moistureSeed := rng.DeriveSeed(worldSeed, "moisture")
	moistureGenerator := NewDSGenerator(g.width, rng.NewRNG(moistureSeed))
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("seed derived for moisture map generation")
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("performing diamond-square algorithm to generate moisture map")
	moistureMap := moistureGenerator.Generate()
	return moistureMap
}

func generateFertilityMap(g *Generator, worldSeed [32]byte, elevationMap *simtypes.LayerMap, moistureMap *simtypes.LayerMap) *simtypes.LayerMap {
	fertilityMap := simtypes.NewLayerMap(g.width, g.height)

	elevationMin, elevationMax := minMax(elevationMap)
	moistureMin, moistureMax := minMax(moistureMap)

	noiseSeed := rng.DeriveSeed(worldSeed, "fertility")
	noiseRNG := rng.NewRNG(noiseSeed)

	for x := 0; x < g.width; x++ {
		for y := 0; y < g.height; y++ {
			elevation := normalize(elevationMap.Get(x, y), elevationMin, elevationMax)
			moisture := normalize(moistureMap.Get(x, y), moistureMin, moistureMax)

			extremeElevation := 2.0 * math.Abs(elevation-0.5) // [0, 1]
			penalty := extremeElevation * extremeElevation * elevationPenaltyWeight

			noise := (noiseRNG.Elevation(nil) - 0.5) * 2.0
			variation := noise * variationAmplitude

			fertility := baseFertility + (moisture * moistureWeight) - penalty + variation
			fertilityMap.Set(x, y, clamp(fertility))
		}
	}
	return fertilityMap
}

func minMax(layer *simtypes.LayerMap) (float64, float64) {
	min := layer.Get(0, 0)
	max := layer.Get(0, 0)

	for x := 0; x < layer.Width; x++ {
		for y := 0; y < layer.Height; y++ {
			value := layer.Get(x, y)
			if value < min {
				min = value
			}
			if value > max {
				max = value
			}
		}
	}

	return min, max
}

func normalize(value, min, max float64) float64 {
	if max == min {
		return 0.0
	}
	return (value - min) / (max - min)
}

func clamp(value float64) float64 {
	if value < 0.0 {
		return 0.0
	}
	if value > 1.0 {
		return 1.0
	}
	return value
}
