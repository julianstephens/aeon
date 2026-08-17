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

type MigrationSummary struct {
	Moved            float64
	SourceCells      int
	DestinationCells int
	AverageDistance  float64
	MaxDistance      int
}

type PopulationPressureSummary struct {
	Under25      int
	Range25to50  int
	Range50to75  int
	Range75to100 int
	Over100      int
}

type PopulationModel struct {
	params        PopulationParameters
	tick          int
	lastMigration MigrationSummary
}

func NewPopulationModel(params PopulationParameters) *PopulationModel {
	return &PopulationModel{params: params}
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

		if !cell.Terrain.IsPassable() || cell.FoodCapacity <= 0 {
			continue
		}

		capacity := int(math.Floor(pm.carryingCapacity(cell)))
		if capacity <= 0 {
			continue
		}

		seedableCells = append(seedableCells, seededCell{
			index:    i,
			weight:   cell.FoodCapacity,
			capacity: capacity,
		})
	}

	seedableCells = selectSeedCells(seedableCells, terrainMap, targetPopulation)
	allocated := distributeSeedPopulation(seedableCells, targetPopulation)
	for _, cell := range seedableCells {
		terrainMap.Cells[cell.index].Population = float64(cell.population)
	}

	return allocated
}

func (pm *PopulationModel) CurrentTick() int {
	return pm.tick
}

func (pm *PopulationModel) MigrationSummary() MigrationSummary {
	return pm.lastMigration
}

func (pm *PopulationModel) PopulationPressureSummary(terrainMap *simtypes.TerrainMap) PopulationPressureSummary {
	summary := PopulationPressureSummary{}
	if terrainMap == nil {
		return summary
	}

	for i := range terrainMap.Cells {
		cell := &terrainMap.Cells[i]
		capacity := pm.carryingCapacity(cell)
		if capacity <= 0 || cell.Population <= 0 {
			continue
		}

		pressure := 0.0
		if capacity > 0 {
			pressure = cell.Population / capacity * 100
		}

		switch {
		case pressure < 25:
			summary.Under25++
		case pressure < 50:
			summary.Range25to50++
		case pressure < 75:
			summary.Range50to75++
		case pressure < 100:
			summary.Range75to100++
		default:
			summary.Over100++
		}
	}
	return summary
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
	pm.applyStarvation(terrainMap)
	pm.applyMigration(terrainMap)

	logger.WithFields(map[string]interface{}{
		"year":       pm.tick,
		"population": pm.TotalPopulation(terrainMap),
	}).Debug("population dynamics simulation completed for year")
}

// carryingCapacity converts the environmental food-capacity score into the
// maximum sustainable population for a cell under this population model.
func (pm *PopulationModel) carryingCapacity(cell *simtypes.TerrainCell) float64 {
	if cell == nil || !cell.Terrain.IsPassable() || cell.FoodCapacity <= 0 || pm.params.MaxCellCapacity <= 0 {
		return 0
	}
	return cell.FoodCapacity * pm.params.MaxCellCapacity
}

func (pm *PopulationModel) growPopulation(terrainMap *simtypes.TerrainMap) {
	for i := range terrainMap.Cells {
		cell := &terrainMap.Cells[i]
		capacity := pm.carryingCapacity(cell)
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

func (pm *PopulationModel) applyStarvation(terrainMap *simtypes.TerrainMap) {
	if pm.params.StarvationRate <= 0 {
		return
	}

	for i := range terrainMap.Cells {
		cell := &terrainMap.Cells[i]
		capacity := pm.carryingCapacity(cell)
		if capacity <= 0 {
			cell.Population = 0
			continue
		}

		excess := cell.Population - capacity
		if excess <= 0 {
			continue
		}

		// Only the population above carrying capacity is exposed to the
		// starvation mortality term. Logistic growth already reduces growth
		// as a population approaches capacity.
		mortality := excess * pm.params.StarvationRate
		cell.Population -= mortality
		if cell.Population < 0 {
			cell.Population = 0
		}
	}
}

type migrationFlow struct {
	from   int
	to     int
	amount float64
}

func (pm *PopulationModel) applyMigration(terrainMap *simtypes.TerrainMap) {
	if pm.params.MigrationRate <= 0 {
		pm.lastMigration = MigrationSummary{}
		return
	}

	flows := make([]migrationFlow, 0)
	for i := range terrainMap.Cells {
		cell := &terrainMap.Cells[i]
		capacity := pm.carryingCapacity(cell)
		if capacity <= 0 {
			continue
		}

		migrationThreshold := capacity * 0.75
		excess := 0.0
		switch {
		case cell.Population > capacity:
			excess = cell.Population - capacity
		case cell.Population > migrationThreshold:
			excess = cell.Population - migrationThreshold
		default:
			continue
		}

		available := excess * pm.params.MigrationRate
		if available <= 0 {
			continue
		}

		neighbors := tmNeighbors(terrainMap, i)
		candidates := make([]int, 0, len(neighbors))
		for _, neighbor := range neighbors {
			neighborCell := &terrainMap.Cells[neighbor]
			neighborCapacity := pm.carryingCapacity(neighborCell)
			room := neighborCapacity - neighborCell.Population
			if room > 0 {
				candidates = append(candidates, neighbor)
			}
		}
		if len(candidates) == 0 {
			continue
		}

		roomPerNeighbor := available / float64(len(candidates))
		for _, neighbor := range candidates {
			neighborCell := &terrainMap.Cells[neighbor]
			room := pm.carryingCapacity(neighborCell) - neighborCell.Population
			amount := minFloat(roomPerNeighbor, room)
			if amount > 0 {
				flows = append(flows, migrationFlow{from: i, to: neighbor, amount: amount})
			}
		}
	}

	pm.lastMigration = summarizeMigration(flows, terrainMap)
	for _, flow := range flows {
		terrainMap.Cells[flow.from].Population -= flow.amount
		terrainMap.Cells[flow.to].Population += flow.amount
	}
}

func summarizeMigration(flows []migrationFlow, terrainMap *simtypes.TerrainMap) MigrationSummary {
	if len(flows) == 0 || terrainMap == nil {
		return MigrationSummary{}
	}

	summary := MigrationSummary{Moved: 0, SourceCells: 0, DestinationCells: 0, AverageDistance: 0, MaxDistance: 0}
	sourceSet := make(map[int]struct{}, len(flows))
	destinationSet := make(map[int]struct{}, len(flows))
	totalDistance := 0.0
	maxDistance := 0
	for _, flow := range flows {
		summary.Moved += flow.amount
		sourceSet[flow.from] = struct{}{}
		destinationSet[flow.to] = struct{}{}

		from := terrainMap.Cells[flow.from].Location
		to := terrainMap.Cells[flow.to].Location
		distance := abs(from.X-to.X) + abs(from.Y-to.Y)
		totalDistance += float64(distance)
		if distance > maxDistance {
			maxDistance = distance
		}
	}

	summary.SourceCells = len(sourceSet)
	summary.DestinationCells = len(destinationSet)
	summary.MaxDistance = maxDistance
	if len(flows) > 0 {
		summary.AverageDistance = totalDistance / float64(len(flows))
	}
	return summary
}

func tmNeighbors(terrainMap *simtypes.TerrainMap, index int) []int {
	x := index % terrainMap.Width
	y := index / terrainMap.Width
	neighbors := make([]int, 0, 4)
	for _, position := range [][2]int{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}} {
		nx, ny := position[0], position[1]
		if nx < 0 || nx >= terrainMap.Width || ny < 0 || ny >= terrainMap.Height {
			continue
		}
		neighbors = append(neighbors, ny*terrainMap.Width+nx)
	}
	return neighbors
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func selectSeedCells(cells []seededCell, terrainMap *simtypes.TerrainMap, population int) []seededCell {
	if len(cells) <= 1 {
		return cells
	}

	seedCount := min(5, max(1, int(math.Ceil(math.Sqrt(float64(population)/10)))))
	seedCount = min(seedCount, len(cells))

	sort.Slice(cells, func(i, j int) bool {
		if cells[i].weight == cells[j].weight {
			return cells[i].index < cells[j].index
		}
		return cells[i].weight > cells[j].weight
	})

	selected := make([]seededCell, 0, seedCount)
	mapSpan := max(terrainMap.Width, terrainMap.Height)
	minDistance := max(1, mapSpan/5)
	for _, candidate := range cells {
		if len(selected) == seedCount {
			break
		}

		position := terrainMap.Cells[candidate.index].Location
		tooClose := false
		for _, chosen := range selected {
			chosenPosition := terrainMap.Cells[chosen.index].Location
			distance := abs(position.X-chosenPosition.X) + abs(position.Y-chosenPosition.Y)
			if distance < minDistance {
				tooClose = true
				break
			}
		}
		if tooClose {
			continue
		}

		selected = append(selected, candidate)
	}

	if len(selected) == 0 {
		selected = append(selected, cells[0])
	}
	return selected
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
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
