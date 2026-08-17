package simulation

import "github.com/julianstephens/aeon/internal/simtypes"

type PopulationMetrics struct {
	Population        float64
	CarryingCapacity  float64
	Utilization       float64
	OccupiedCells     int
	OverCapacityCells int
	Pressure          PopulationPressureSummary
	ByTerrain         []TerrainPopulationMetrics
}

type TerrainPopulationMetrics struct {
	Terrain          string
	Population       float64
	CarryingCapacity float64
	Utilization      float64
	OccupiedCapacity float64
	OccupiedRatio    float64
}

func CollectPopulationMetrics(terrainMap *simtypes.TerrainMap, pm *PopulationModel) PopulationMetrics {
	metrics := PopulationMetrics{}
	if terrainMap == nil || pm == nil {
		return metrics
	}

	metrics.ByTerrain = []TerrainPopulationMetrics{
		{Terrain: simtypes.TerrainTypePlains.String()},
		{Terrain: simtypes.TerrainTypeForest.String()},
		{Terrain: simtypes.TerrainTypeMountain.String()},
		{Terrain: simtypes.TerrainTypeWater.String()},
	}

	for i := range terrainMap.Cells {
		cell := &terrainMap.Cells[i]
		cellCapacity := pm.carryingCapacity(cell)
		metrics.Population += cell.Population
		metrics.CarryingCapacity += cellCapacity

		if cell.Terrain >= simtypes.TerrainTypePlains && cell.Terrain <= simtypes.TerrainTypeWater {
			idx := int(cell.Terrain)
			metrics.ByTerrain[idx].Population += cell.Population
			metrics.ByTerrain[idx].CarryingCapacity += cellCapacity
			if cell.Population > 0 {
				metrics.ByTerrain[idx].OccupiedCapacity += cellCapacity
			}
		}

		if cell.Population <= 0 {
			continue
		}

		metrics.OccupiedCells++
		if cell.Population > cellCapacity {
			metrics.OverCapacityCells++
		}

		if cellCapacity <= 0 {
			continue
		}

		pressure := cell.Population / cellCapacity * 100
		switch {
		case pressure < 25:
			metrics.Pressure.Under25++
		case pressure < 50:
			metrics.Pressure.Range25to50++
		case pressure < 75:
			metrics.Pressure.Range50to75++
		case pressure < 100:
			metrics.Pressure.Range75to100++
		default:
			metrics.Pressure.Over100++
		}
	}

	if metrics.CarryingCapacity > 0 {
		metrics.Utilization = metrics.Population / metrics.CarryingCapacity
	}

	for i := range metrics.ByTerrain {
		if metrics.ByTerrain[i].CarryingCapacity > 0 {
			metrics.ByTerrain[i].Utilization =
				metrics.ByTerrain[i].Population / metrics.ByTerrain[i].CarryingCapacity
			metrics.ByTerrain[i].OccupiedRatio =
				metrics.ByTerrain[i].OccupiedCapacity / metrics.ByTerrain[i].CarryingCapacity
		}
	}

	return metrics
}
