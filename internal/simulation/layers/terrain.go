package layers

import (
	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/go-utils/logger"
)

const (
	MaxClassificationIterations = 3
)

type TerrainClassifier struct {
	tm *simtypes.TerrainMap
}

func NewTerrainClassifier(tm *simtypes.TerrainMap) *TerrainClassifier {
	return &TerrainClassifier{
		tm: tm,
	}
}

// Classify takes a terrain map and classifies each cell into a terrain type based on elevation, moisture, and fertility.
func (tc *TerrainClassifier) Classify(elevation, moisture, fertility, terrain *simtypes.Layer) error {
	logger.Debug("building initial terrain layer")
	if err := tc.setInitialTerrainTypes(terrain); err != nil {
		return err
	}
	logger.Debug("initial terrain layer built")

	logger.WithFields(map[string]interface{}{
		"max_iterations": MaxClassificationIterations,
	}).Debug("beginning terrain classification")
	for i := range MaxClassificationIterations {
		for x := 0; x < tc.tm.Width; x++ {
			for y := 0; y < tc.tm.Height; y++ {
				cell := tc.tm.GetCell(x, y)
				elev := elevation.Get(x, y)
				moist := moisture.Get(x, y)
				fert := fertility.Get(x, y)
				ter := terrain.Get(x, y)

				terrainType := tc.classifyCell(*cell, elev, moist, fert, ter, tc.getNeighboringTerrainTypes(*cell))

				if err := tc.tm.SetTerrainType(x, y, terrainType); err != nil {
					return &PipelineError{
						Code:    CodeClassificationError,
						Message: "Failed to set terrain type",
						Cause:   err,
					}
				}
			}
		}
		logger.WithFields(map[string]any{
			"iteration": i + 1,
		}).Debug("terrain classification iteration completed")
	}
	logger.Debug("terrain classification completed")
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

func (tc *TerrainClassifier) getNeighboringTerrainTypes(cell simtypes.TerrainCell) map[simtypes.TerrainType]int {
	neighboringTypes := make(map[simtypes.TerrainType]int)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			neighborX := cell.Location.X + dx
			neighborY := cell.Location.Y + dy
			neighborCell := tc.tm.GetCell(neighborX, neighborY)
			if neighborCell != nil {
				neighboringTypes[neighborCell.Terrain]++
			}
		}
	}

	return neighboringTypes
}

func (tc *TerrainClassifier) classifyCell(
	cell simtypes.TerrainCell,
	elevation, moisture, fertility, terrain float64,
	neighboringTypes map[simtypes.TerrainType]int,
) simtypes.TerrainType {
	// 1. If the cell is an isolated forest surrounded by plains, it becomes plains.
	if cell.Terrain == simtypes.TerrainTypeForest && isIsolatedForest(neighboringTypes) {
		return simtypes.TerrainTypePlains
	}

	// 2. If the cell is an isolated plain surrounded by forest, it becomes forest.
	if cell.Terrain == simtypes.TerrainTypePlains && isIsolatedPlains(neighboringTypes) {
		return simtypes.TerrainTypeForest
	}

	// 3. If the cell is a isolated water surrounded by land, it becomes land.
	isolated, isolatedBy := isIsolatedWater(neighboringTypes)
	if cell.Terrain == simtypes.TerrainTypeWater && isolated {
		return isolatedBy
	}

	// 4. If the cell is water, it remains water.
	if cell.Terrain == simtypes.TerrainTypeWater || elevation < WaterThreshold {
		return simtypes.TerrainTypeWater
	}

	// 5. If the cell is a mountain, it remains a mountain.
	if cell.Terrain == simtypes.TerrainTypeMountain || elevation > MountainThreshold {
		return simtypes.TerrainTypeMountain
	}

	return cell.Terrain
}

func isIsolatedForest(neighboringTypes map[simtypes.TerrainType]int) bool {
	if neighboringTypes[simtypes.TerrainTypeForest] > 0 {
		return false
	}
	return true
}

func isIsolatedPlains(neighboringTypes map[simtypes.TerrainType]int) bool {
	if neighboringTypes[simtypes.TerrainTypePlains] > 0 {
		return false
	}
	return true
}

func isIsolatedWater(neighboringTypes map[simtypes.TerrainType]int) (bool, simtypes.TerrainType) {
	if neighboringTypes[simtypes.TerrainTypeWater] > 0 {
		return false, simtypes.TerrainTypeWater
	}
	maxCount := 0
	newTerrainType := simtypes.TerrainTypePlains
	for terrainType, count := range neighboringTypes {
		if count > maxCount && terrainType != simtypes.TerrainTypeWater {
			maxCount = count
			newTerrainType = terrainType
		}
	}
	return true, newTerrainType
}

type TerrainSmoother struct{}

// Smooth takes a terrain map and applies smoothing to the terrain types to create more natural transitions.
func (ts *TerrainSmoother) Smooth(terrainMap *simtypes.TerrainMap) *simtypes.TerrainMap {
	return nil
}
