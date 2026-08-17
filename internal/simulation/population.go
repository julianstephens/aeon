package simulation

import (
	"math"
	"sort"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/go-utils/logger"
)

type PopulationParameters struct {
	GrowthRate      float64
	StarvationRate  float64
	MigrationRate   float64
	MaxCellCapacity float64
}

type PopulationModel struct {
	params PopulationParameters
	tick   int
}

func NewPopulationModel(params PopulationParameters) *PopulationModel {
	return &PopulationModel{
		params: params,
	}
}

func (pm *PopulationModel) SimulatePopulationDynamics(durationYears int, terrainMap *simtypes.TerrainMap) {
	for range durationYears {
		pm.AdvanceOneYear(terrainMap)
	}
}

func (pm *PopulationModel) AdvanceOneYear(terrainMap *simtypes.TerrainMap) {
	pm.simulateYear(terrainMap)
	pm.tick++
}

type seededCell struct {
	index      int
	weight     float64
	capacity   int
	population int
}

type seededCellRemainder struct {
	cellIndex int
	frac      float64
}

func (pm *PopulationModel) SeedPopulation(terrainMap *simtypes.TerrainMap, population float64) int {
	if terrainMap == nil || population <= 0 {
		return 0
	}

	targetPopulation := int(math.Round(population))
	if targetPopulation <= 0 {
		return 0
	}

	seedableCells := make([]seededCell, 0, len(terrainMap.Cells))
	for i := range terrainMap.Cells {
		cell := &terrainMap.Cells[i]
		cell.Population = 0

		if cell.FoodCapacity <= 0 {
			continue
		}

		capacity := int(math.Floor(cell.FoodCapacity * pm.params.MaxCellCapacity))
		if capacity <= 0 {
			continue
		}

		seedableCells = append(seedableCells, seededCell{
			index:      i,
			weight:     cell.FoodCapacity,
			capacity:   capacity,
			population: 0,
		})
	}

	allocated := distributeSeedPopulation(seedableCells, targetPopulation)
	for _, cell := range seedableCells {
		terrainMap.Cells[cell.index].Population = float64(cell.population)
	}

	return allocated
}

func (pm *PopulationModel) CurrentTick() int {
	return pm.tick
}

func (pm *PopulationModel) TotalPopulation(terrainMap *simtypes.TerrainMap) int {
	var total float64
	for i := range terrainMap.Cells {
		total += terrainMap.Cells[i].Population
	}
	return int(math.Round(total))
}

func (pm *PopulationModel) simulateYear(terrainMap *simtypes.TerrainMap) {
	logger.WithFields(map[string]interface{}{
		"year":       pm.tick,
		"population": pm.TotalPopulation(terrainMap),
	}).Debug("simulating population dynamics for year")
	pm.growPopulation(terrainMap)
	logger.WithFields(map[string]interface{}{
		"year":       pm.tick,
		"population": pm.TotalPopulation(terrainMap),
	}).Debug("population dynamics simulation completed for year")
}

func (pm *PopulationModel) growPopulation(terrainMap *simtypes.TerrainMap) {
	for i := range terrainMap.Cells {
		cell := &terrainMap.Cells[i]
		capacity := cell.FoodCapacity * pm.params.MaxCellCapacity
		if capacity <= 0 {
			cell.Population = 0
			continue
		}

		growth := cell.Population * pm.params.GrowthRate * (1 - cell.Population/capacity)
		cell.Population += growth

		if cell.Population < 0 {
			cell.Population = 0
		}
	}
}

func distributeSeedPopulation(cells []seededCell, targetPopulation int) int {
	remaining := targetPopulation
	active := make([]int, 0, len(cells))
	for i := range cells {
		if cells[i].capacity > 0 && cells[i].weight > 0 {
			active = append(active, i)
		}
	}

	for remaining > 0 && len(active) > 0 {
		roundTarget := remaining

		var totalWeight float64
		for _, idx := range active {
			totalWeight += cells[idx].weight
		}
		if totalWeight <= 0 {
			break
		}

		assignedThisRound := 0
		remainders := make([]seededCellRemainder, 0, len(active))
		for _, idx := range active {
			cell := &cells[idx]
			room := cell.capacity - cell.population
			if room <= 0 {
				continue
			}

			idealShare := float64(roundTarget) * (cell.weight / totalWeight)
			base := min(int(math.Floor(idealShare)), room)

			if base > 0 {
				cell.population += base
				remaining -= base
				assignedThisRound += base
			}

			remainders = append(remainders, seededCellRemainder{
				cellIndex: idx,
				frac:      idealShare - math.Floor(idealShare),
			})
		}

		if remaining > 0 {
			sort.Slice(remainders, func(i, j int) bool {
				if remainders[i].frac == remainders[j].frac {
					return remainders[i].cellIndex < remainders[j].cellIndex
				}
				return remainders[i].frac > remainders[j].frac
			})

			for _, rem := range remainders {
				if remaining <= 0 {
					break
				}

				cell := &cells[rem.cellIndex]
				if cell.population >= cell.capacity {
					continue
				}

				cell.population++
				remaining--
				assignedThisRound++
			}
		}

		nextActive := active[:0]
		for _, idx := range active {
			if cells[idx].population < cells[idx].capacity {
				nextActive = append(nextActive, idx)
			}
		}
		active = nextActive

		if assignedThisRound == 0 {
			break
		}
	}

	return targetPopulation - remaining
}
