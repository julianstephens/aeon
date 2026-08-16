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

	WaterThreshold    = 0.35
	MountainThreshold = 0.85
	ForestMoisture    = 0.60

	WaterTarget    = 0.35
	PlainsTarget   = 0.35
	ForestTarget   = 0.22
	MountainTarget = 0.08

	MinWaterRegionSize = 8

	MinSettlementDistance = 12

	PlainsFoodCapacity   = 100.0
	ForestFoodCapacity   = 75.0
	MountainFoodCapacity = 15.0
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
	// generate elevation layer using Diamond-Square algorithm
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("generating elevation layer")
	elevationLayer := generateElevationLayer(g, worldSeed)
	logger.WithFields(map[string]interface{}{
		"width":  elevationLayer.Width,
		"height": elevationLayer.Height,
	}).Debug("elevation layer generated")

	// generate moisture layer using Diamond-Square algorithm
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("generating moisture layer")
	moistureLayer := generateMoistureLayer(g, worldSeed)
	logger.WithFields(map[string]interface{}{
		"width":  moistureLayer.Width,
		"height": moistureLayer.Height,
	}).Debug("moisture layer generated")

	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("generating fertility layer")
	fertilityLayer := generateFertilityLayer(g, worldSeed, elevationLayer, moistureLayer)
	logger.WithFields(map[string]interface{}{
		"width":  fertilityLayer.Width,
		"height": fertilityLayer.Height,
	}).Debug("fertility layer generated")

	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("generating terrain layer")
	terrainLayer := generateTerrainLayer(elevationLayer, moistureLayer, fertilityLayer)
	logger.WithFields(map[string]interface{}{
		"width":  terrainLayer.Width,
		"height": terrainLayer.Height,
	}).Debug("terrain layer generated")

	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("applying elevation layer to terrain map")
	tm.ApplyElevation(elevationLayer)
	logger.Debug("elevation layer applied to terrain map")
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("applying moisture layer to terrain map")
	tm.ApplyMoisture(moistureLayer)
	logger.Debug("moisture layer applied to terrain map")
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("applying fertility layer to terrain map")
	tm.ApplyFertility(fertilityLayer)
	logger.Debug("fertility layer applied to terrain map")
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("applying terrain layer to terrain map")
	tm.ApplyTerrain(terrainLayer)
	logger.Debug("terrain layer applied to terrain map")

	tm.SetInitialized(true)
	return nil
}

func generateElevationLayer(g *Generator, worldSeed [32]byte) *simtypes.LayerMap {
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("deriving elevation seed from world seed")
	elevationSeed := rng.DeriveSeed(worldSeed, "elevation")
	elevationGenerator := NewDSGenerator(g.width, rng.NewRNG(elevationSeed))
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("seed derived for elevation layer generation")
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("performing diamond-square algorithm to generate elevation layer")
	elevationLayer := elevationGenerator.Generate()
	return elevationLayer
}

func generateMoistureLayer(g *Generator, worldSeed [32]byte) *simtypes.LayerMap {
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("deriving moisture seed from world seed")
	moistureSeed := rng.DeriveSeed(worldSeed, "moisture")
	moistureGenerator := NewDSGenerator(g.width, rng.NewRNG(moistureSeed))
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("seed derived for moisture layer generation")
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("performing diamond-square algorithm to generate moisture layer")
	moistureLayer := moistureGenerator.Generate()
	return moistureLayer
}

func generateFertilityLayer(
	g *Generator,
	worldSeed [32]byte,
	elevationMap *simtypes.LayerMap,
	moistureMap *simtypes.LayerMap,
) *simtypes.LayerMap {
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

func generateTerrainLayer(elevationMap, moistureMap, _fertilityMap *simtypes.LayerMap) *simtypes.LayerMap {
	terrainLayer := simtypes.NewLayerMap(elevationMap.Width, elevationMap.Height)

	for x := 0; x < elevationMap.Width; x++ {
		for y := 0; y < elevationMap.Height; y++ {
			elevation := elevationMap.Get(x, y)
			moisture := moistureMap.Get(x, y)

			if elevation < WaterThreshold {
				terrainLayer.Set(x, y, float64(simtypes.TerrainTypeWater))
			} else if elevation > MountainThreshold {
				terrainLayer.Set(x, y, float64(simtypes.TerrainTypeMountain))
			} else if moisture > ForestMoisture {
				terrainLayer.Set(x, y, float64(simtypes.TerrainTypeForest))
			} else {
				terrainLayer.Set(x, y, float64(simtypes.TerrainTypePlains))
			}
		}
	}

	return terrainLayer
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
