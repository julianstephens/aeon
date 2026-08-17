package simulation

import (
	"fmt"
	"math"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/go-utils/cliutil"
	"github.com/julianstephens/go-utils/logger"
)

const (
	InitialPopulationCount = 100
	ReportIntervalYears    = 25
)

// Run runs the simulation for a given number of years and prints the results to the console.
func Run(c *cliutil.Console, seed string, years int) error {
	logger.WithFields(map[string]any{
		"seed":  seed,
		"years": years,
	}).Debug("starting simulation run. creating world")

	populationParams := PopulationParameters{
		GrowthRate:      0.025,
		StarvationRate:  0.10,
		MigrationRate:   0.07,
		MaxCellCapacity: 100,
	}

	w, err := NewWorld(seed, populationParams)
	if err != nil {
		logger.Errorf("failed to initialize world: %v", err)
		return err
	}
	logger.WithFields(map[string]any{
		"year":       w.Year(),
		"population": w.PopulationCount(),
	}).Debug("world initialized")
	if c != nil {
		printYearlyReport(w.Year(), w.terrainMap, w.populationModel)
	}

	logger.WithFields(map[string]any{
		"year":       w.Year(),
		"population": w.PopulationCount(),
	}).Debug("starting simulation run")
	for range years {
		w.populationModel.AdvanceOneYear(w.terrainMap)
		if c != nil && shouldReportYear(w.Year(), years) {
			println("")
			printYearlyReport(w.Year(), w.terrainMap, w.populationModel)
		}
	}
	logger.Debug("simulation run completed")
	if c != nil {
		_ = c.Success(fmt.Sprintf("Simulation completed after %d years", years))
	}
	return nil
}

func shouldReportYear(year, totalYears int) bool {
	if year == 0 || year == totalYears {
		return true
	}
	return year%ReportIntervalYears == 0
}

func printYearlyReport(year int, terrainMap *simtypes.TerrainMap, populationModel *PopulationModel) {
	stats := summarizeTerrainPopulation(terrainMap, populationModel)
	migration := populationModel.MigrationSummary()
	pressure := populationModel.PopulationPressureSummary(terrainMap)
	println(fmt.Sprintf("Year %d", year))
	println("Population")
	println(fmt.Sprintf("  total:              %d", stats.totalPopulation))
	println(fmt.Sprintf("  carrying capacity:  %d", stats.carryingCapacity))
	println(fmt.Sprintf("  utilization:        %.1f%%", stats.utilizationPercent))
	println(fmt.Sprintf("  occupied cells:     %d", stats.occupiedCells))
	println(fmt.Sprintf("  over-capacity cells: %d", stats.overCapacityCells))
	println()
	println("Migration")
	println(fmt.Sprintf("  moved:             %d", int(math.Round(migration.Moved))))
	println(fmt.Sprintf("  source cells:      %d", migration.SourceCells))
	println(fmt.Sprintf("  destination cells: %d", migration.DestinationCells))
	println(fmt.Sprintf("  average distance:  %.1f", migration.AverageDistance))
	println(fmt.Sprintf("  max distance:      %d", migration.MaxDistance))
	println()
	println("Population pressure")
	println(fmt.Sprintf("  under 25%%:         %d cells", pressure.Under25))
	println(fmt.Sprintf("  25–50%%:           %d cells", pressure.Range25to50))
	println(fmt.Sprintf("  50–75%%:           %d cells", pressure.Range50to75))
	println(fmt.Sprintf("  75–100%%:          %d cells", pressure.Range75to100))
	println(fmt.Sprintf("  >100%%:            %d cells", pressure.Over100))
}

type terrainPopulationSummary struct {
	totalPopulation    int
	carryingCapacity   int
	utilizationPercent float64
	occupiedCells      int
	overCapacityCells  int
}

func summarizeTerrainPopulation(
	terrainMap *simtypes.TerrainMap,
	populationModel *PopulationModel,
) terrainPopulationSummary {
	if terrainMap == nil || populationModel == nil {
		return terrainPopulationSummary{}
	}

	var totalPopulation float64
	var totalCarryingCapacity float64
	occupiedCells := 0
	overCapacityCells := 0

	for i := range terrainMap.Cells {
		cell := &terrainMap.Cells[i]
		cellCapacity := populationModel.carryingCapacity(cell)
		totalPopulation += cell.Population
		totalCarryingCapacity += cellCapacity
		if cell.Population > 0 {
			occupiedCells++
		}
		if cell.Population > cellCapacity {
			overCapacityCells++
		}
	}

	utilizationPercent := 0.0
	if totalCarryingCapacity > 0 {
		utilizationPercent = (totalPopulation / totalCarryingCapacity) * 100
	}

	return terrainPopulationSummary{
		totalPopulation:    int(math.Round(totalPopulation)),
		carryingCapacity:   int(math.Round(totalCarryingCapacity)),
		utilizationPercent: utilizationPercent,
		occupiedCells:      occupiedCells,
		overCapacityCells:  overCapacityCells,
	}
}
