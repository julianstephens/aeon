package layers

import (
	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/go-utils/logger"
)

const (
	// Elevation values at these extremes are treated as hard geographic constraints.
	WaterElevationThreshold    = 0.12
	MountainElevationThreshold = 0.92

	// Scores are weighted independently so the middle of the elevation range can
	// be classified using moisture, fertility, and neighborhood context.
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
	return &TerrainClassifier{
		tm: tm,
	}
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
					cell.Location,
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
	// Terrain classification owns the TerrainType field, so initialize it from
	// the generated terrain layer regardless of the TerrainMap's initialized flag.
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
	location simtypes.Position,
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
	return TerrainScores{
		Water: waterScore(elevation, moisture, neighbors.fraction(neighbors.Water)),
		Plains: plainsScore(
			elevation,
			moisture,
			fertility,
			neighbors.fraction(neighbors.Plains),
		),
		Forest: forestScore(
			elevation,
			moisture,
			fertility,
			neighbors.fraction(neighbors.Forest),
		),
		Mountain: mountainScore(
			elevation,
			fertility,
			neighbors.fraction(neighbors.Mountain),
		),
	}
}

func waterScore(elevation, moisture, waterNeighbors float64) float64 {
	lowland := 1 - normalizeRange(elevation, WaterElevationThreshold, MountainElevationThreshold)
	return WaterElevationWeight*lowland +
		WaterMoistureWeight*moisture +
		WaterNeighborhoodWeight*waterNeighbors
}

func mountainScore(elevation, fertility, mountainNeighbors float64) float64 {
	highland := normalizeRange(elevation, WaterElevationThreshold, MountainElevationThreshold)
	fertilityPenalty := fertility * MountainFertilityPenalty
	return MountainElevationWeight*highland +
		MountainNeighborhoodWeight*mountainNeighbors +
		(1-fertilityPenalty)
}

func forestScore(elevation, moisture, fertility, forestNeighbors float64) float64 {
	moderateElevation := 1 - abs(2*elevation-0.75)
	return ForestMoistureWeight*moisture +
		ForestFertilityWeight*fertility +
		ForestElevationWeight*clamp01(moderateElevation) +
		ForestNeighborhoodWeight*forestNeighbors
}

func plainsScore(elevation, moisture, fertility, plainsNeighbors float64) float64 {
	moderateElevation := 1 - abs(2*elevation-0.55)
	moderateMoisture := 1 - abs(2*moisture-0.75)
	return PlainsFertilityWeight*fertility +
		PlainsMoistureWeight*clamp01(moderateMoisture) +
		PlainsElevationWeight*clamp01(moderateElevation) +
		PlainsNeighborhoodWeight*plainsNeighbors
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
