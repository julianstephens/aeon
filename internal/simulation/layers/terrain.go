package layers

import (
	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/go-utils/logger"
)

const (
	WaterElevationThreshold    = 0.10
	MountainElevationThreshold = 0.90

	WaterElevationWeight    = 0.65
	WaterMoistureWeight     = 0.25
	WaterNeighborhoodWeight = 0.10

	MountainElevationWeight    = 0.70
	MountainFertilityPenalty   = 0.15
	MountainNeighborhoodWeight = 0.15

	ForestMoistureWeight     = 0.55
	ForestFertilityWeight    = 0.25
	ForestElevationWeight    = 0.15
	ForestNeighborhoodWeight = 0.05

	PlainsFertilityWeight    = 0.40
	PlainsMoistureWeight     = 0.20
	PlainsElevationWeight    = 0.35
	PlainsNeighborhoodWeight = 0.05

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

// Classify assigns terrain synchronously from the scalar layers and the
// previous iteration's terrain state. Physical fields dominate the score;
// neighborhood terms provide only weak spatial reinforcement.
func (tc *TerrainClassifier) Classify(elevation, moisture, fertility *simtypes.Layer) error {
	logger.Debug("building initial terrain layer")
	if err := tc.seedInitialTerrain(elevation, moisture, fertility); err != nil {
		return err
	}

	for i := 0; i < MaxClassificationIterations; i++ {
		next := simtypes.NewLayer(tc.tm.Width, tc.tm.Height)

		for x := 0; x < tc.tm.Width; x++ {
			for y := 0; y < tc.tm.Height; y++ {
				neighbors := tc.getNeighborTerrainStats(tc.tm, x, y)
				terrainType := tc.classifyCell(
					elevation.Get(x, y),
					moisture.Get(x, y),
					fertility.Get(x, y),
					neighbors,
				)
				next.Set(x, y, float64(terrainType))
			}
		}

		tc.tm.ApplyTerrain(next)
	}

	return nil
}

func (tc *TerrainClassifier) seedInitialTerrain(elevation, moisture, fertility *simtypes.Layer) error {
	for x := 0; x < tc.tm.Width; x++ {
		for y := 0; y < tc.tm.Height; y++ {
			terrain := tc.classifyCell(
				elevation.Get(x, y),
				moisture.Get(x, y),
				fertility.Get(x, y),
				NeighborTerrainStats{},
			)
			if err := tc.tm.SetTerrainType(x, y, terrain); err != nil {
				return &PipelineError{
					Code:    CodeClassificationError,
					Message: "Failed to seed initial terrain type",
					Cause:   err,
				}
			}
		}
	}
	return nil
}

func (tc *TerrainClassifier) getNeighborTerrainStats(tm *simtypes.TerrainMap, x, y int) NeighborTerrainStats {
	var stats NeighborTerrainStats
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			cell := tm.GetCell(x+dx, y+dy)
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
	return TerrainScores{
		Water: waterScore(
			elevation,
			moisture,
			neighbors.fraction(neighbors.Water),
		),
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
		WaterMoistureWeight*clamp01(moisture) +
		WaterNeighborhoodWeight*clamp01(waterNeighbors)
}

func mountainScore(elevation, fertility, mountainNeighbors float64) float64 {
	highland := normalizeRange(elevation, WaterElevationThreshold, MountainElevationThreshold)
	return MountainElevationWeight*highland +
		MountainNeighborhoodWeight*clamp01(mountainNeighbors) -
		MountainFertilityPenalty*clamp01(fertility)
}

func forestScore(elevation, moisture, fertility, forestNeighbors float64) float64 {
	elevationPreference := 1 - abs(2*elevation-0.65)
	return ForestMoistureWeight*clamp01(moisture) +
		ForestFertilityWeight*clamp01(fertility) +
		ForestElevationWeight*clamp01(elevationPreference) +
		ForestNeighborhoodWeight*clamp01(forestNeighbors)
}

func plainsScore(elevation, moisture, fertility, plainsNeighbors float64) float64 {
	elevationPreference := 1 - abs(2*elevation-0.50)
	moisturePreference := 1 - abs(2*moisture-0.45)
	return PlainsFertilityWeight*clamp01(fertility) +
		PlainsMoistureWeight*clamp01(moisturePreference) +
		PlainsElevationWeight*clamp01(elevationPreference) +
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
