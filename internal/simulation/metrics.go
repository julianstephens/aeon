package simulation

import "github.com/julianstephens/aeon/internal/simtypes"

type PopulationMetrics struct {
	Population        float64
	CarryingCapacity  float64
	Utilization       float64
	OccupiedCells     int
	OverCapacityCells int
	Pressure          PopulationPressureSummary
}

func CollectPopulationMetrics(terrainMap *simtypes.TerrainMap, pm *PopulationModel) PopulationMetrics {
	metrics := PopulationMetrics{}
	if terrainMap == nil || pm == nil {
		return metrics
	}

	for i := range terrainMap.Cells {
		cell := &terrainMap.Cells[i]
		cellCapacity := pm.carryingCapacity(cell)
		metrics.Population += cell.Population
		metrics.CarryingCapacity += cellCapacity

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

	return metrics
}
