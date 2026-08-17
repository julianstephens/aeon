package layers

import (
	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/go-utils/logger"
)

const (
	// Elevation values at these extremes are treated as hard geographic constraints.
	WaterElevationThreshold    = 0.12
	MountainElevationThreshold = 0.92

	// Scores are weighted so each terrain type competes on the same approximate [0, 1] scale.
	WaterElevationWeight       = 0.55
	WaterMoistureWeight        = 0.20
	WaterNeighborhoodWeight    = 0.25
	MountainElevationWeight    = 0.60
	MountainNeighborhoodWeight = 0.25
	MountainFertilityPenalty   = 0.15
	ForestMoistureWeight       = 0.50
	ForestFertilityWeight      = 0.20
	ForestElevationWeight      = 0.10
	ForestNeighborhoodWeight   = 0.20
	PlainsFertilityWeight      = 0.35
	PlainsMoistureWeight       = 0.20
	PlainsElevationWeight      = 0.25
	PlainsNeighborhoodWeight   = 0.20

	MaxClassificationIterations = 3
)

type TerrainClassifier struct {
	tm *simtypes.TerrainMap
}

type TerrainScores struct {
	Water    float64
	Plains   float64
	Forest   float64
	Mountain float64
}

type NeighborTerrainStats struct {
	Total    int
	Water    int
	Plains   int
	Forest   int
	Mountain int
}

func NewTerrainClassifier(tm *simtypes.TerrainMap) *TerrainClassifier {
	return &TerrainClassifier{tm: tm}
}

// Classify assigns a terrain type to every cell using elevation, moisture,
// fertility, and local terrain context. Elevation provides hard constraints
// only at geographic extremes; the middle of the range is resolved by scores.
func (tc *TerrainClassifier) Classify(elevation, moisture, fertility, terrain *simtypes.Layer) error {
	logger.Debug("building initial terrain layer")
	if err := tc.setInitialTerrainTypes(terrain); err != nil {
		return err
	}

	for i := 0; i < MaxClassificationIterations; i++ {
		for x := 0; x < tc.tm.Width; x++ {
			for y := 0; y < tc.tm.Height; y++ {
				cell := tc.tm.GetCell(x, y)
				if cell == nil {
					continue
				}

				terrainType := tc.classifyCell(
					elevation.Get(x, y),
					moisture.Get(x, y),
					fertility.Get(x, y),
					tc.getNeighborTerrainStats(cell.Location),
				)

				if err := tc.tm.SetTerrainType(x, y, terrainType); err != nil {
					return &PipelineError{
						Code:    CodeClassificationError,
						Message: "Failed to set terrain type",
						Cause:   err,
					}
				}
			}
		}
	}

	return nil
}

func (tc *TerrainClassifier) setInitialTerrainTypes(terrain *simtypes.Layer) error {
	if tc.tm.IsInitialized() {
		return nil
	}

	for x := 0; x < tc.tm.Width; x++ {
		for y := 0; y < tc.tm.Height; y++ {
			ter := terrain.Get(x, y)
			if err := tc.tm.SetTerrainType(x, y, simtypes.TerrainType(ter)); err != nil {
				return &PipelineError{
					Code:    CodeClassificationError,
					Message: "Failed to set initial terrain type",
					Cause:   err,
				}
			}
		}
	}

	tc.tm.SetInitialized(true)
	return nil
}

func (tc *TerrainClassifier) getNeighborTerrainStats(position simtypes.Position) NeighborTerrainStats {
	var stats NeighborTerrainStats

	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}

			cell := tc.tm.GetCell(position.X+dx, position.Y+dy)
			if cell == nil {
				continue
			}

			stats.Total++
			switch cell.Terrain {
			case simtypes.TerrainTypeWater:
				stats.Water++
			case simtypes.TerrainTypePlains:
				stats.Plains++
			case simtypes.TerrainTypeForest:
				stats.Forest++
			case simtypes.TerrainTypeMountain:
				stats.Mountain++
			}
		}
	}

	return stats
}

func (s NeighborTerrainStats) fraction(count int) float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(count) / float64(s.Total)
}

func (tc *TerrainClassifier) classifyCell(
	elevation, moisture, fertility float64,
	neighbors NeighborTerrainStats,
) simtypes.TerrainType {
	if elevation <= WaterElevationThreshold {
		return simtypes.TerrainTypeWater
	}
	if elevation >= MountainElevationThreshold {
		return simtypes.TerrainTypeMountain
	}

	scores := tc.scoreCell(elevation, moisture, fertility, neighbors)
	return scores.maxTerrain()
}

func (tc *TerrainClassifier) scoreCell(
	elevation, moisture, fertility float64,
	neighbors NeighborTerrainStats,
) TerrainScores {
	waterNeighbors := neighbors.fraction(neighbors.Water)
	mountainNeighbors := neighbors.fraction(neighbors.Mountain)
	forestNeighbors := neighbors.fraction(neighbors.Forest)
	plainsNeighbors := neighbors.fraction(neighbors.Plains)

	return TerrainScores{
		Water: waterScore(elevation, moisture, waterNeighbors),
		Plains: plainsScore(
			elevation,
			moisture,
			fertility,
			plainsNeighbors,
		),
		Forest: forestScore(
			elevation,
			moisture,
			fertility,
			forestNeighbors,
		),
		Mountain: mountainScore(
			elevation,
			fertility,
			mountainNeighbors,
		),
	}
}

func waterScore(elevation, moisture, waterNeighbors float64) float64 {
	lowland := 1 - normalizeRange(elevation, WaterElevationThreshold, MountainElevationThreshold)
	return WaterElevationWeight*lowland +
		WaterMoistureWeight*clamp01(moisture) +
		WaterNeighborhoodWeight*clamp01(waterNeighbors)
}

func mountainScore(elevation, fertility, mountainNeighbors float64) float64 {
	highland := normalizeRange(elevation, WaterElevationThreshold, MountainElevationThreshold)
	fertilityPenalty := clamp01(fertility) * MountainFertilityPenalty
	return MountainElevationWeight*highland +
		MountainNeighborhoodWeight*clamp01(mountainNeighbors) +
		MountainFertilityPenalty*0 +
		(1 - MountainFertilityPenalty) * (1 - fertilityPenalty)
}

func forestScore(elevation, moisture, fertility, forestNeighbors float64) float64 {
	moderateElevation := 1 - abs(2*elevation-0.75)
	return ForestMoistureWeight*clamp01(moisture) +
		ForestFertilityWeight*clamp01(fertility) +
		ForestElevationWeight*clamp01(moderateElevation) +
		ForestNeighborhoodWeight*clamp01(forestNeighbors)
}

func plainsScore(elevation, moisture, fertility, plainsNeighbors float64) float64 {
	moderateElevation := 1 - abs(2*elevation-0.55)
	moderateMoisture := 1 - abs(2*moisture-0.75)
	return PlainsFertilityWeight*clamp01(fertility) +
		PlainsMoistureWeight*clamp01(moderateMoisture) +
		PlainsElevationWeight*clamp01(moderateElevation) +
		PlainsNeighborhoodWeight*clamp01(plainsNeighbors)
}

func (s TerrainScores) maxTerrain() simtypes.TerrainType {
	terrain := simtypes.TerrainTypePlains
	maxScore := s.Plains

	if s.Water > maxScore {
		terrain = simtypes.TerrainTypeWater
		maxScore = s.Water
	}
	if s.Forest > maxScore {
		terrain = simtypes.TerrainTypeForest
		maxScore = s.Forest
	}
	if s.Mountain > maxScore {
		terrain = simtypes.TerrainTypeMountain
	}

	return terrain
}

func normalizeRange(value, minValue, maxValue float64) float64 {
	if maxValue <= minValue {
		return 0
	}
	return clamp01((value - minValue) / (maxValue - minValue))
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
