package simulation_test

import (
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation"
)

func TestPopulationModel_SeedPopulation_WeightsSelectedCentersByFoodCapacity(t *testing.T) {
	pm := simulation.NewPopulationModel(simulation.PopulationParameters{MaxCellCapacity: 100})
	tm := terrainMapWithFoodCapacities(10, 1, 9, 1, 1, 1, 1, 1, 1, 1)

	allocated := pm.SeedPopulation(tm, 20)
	if allocated != 20 {
		t.Fatalf("unexpected seeded total: got %d, want %d", allocated, 20)
	}

	if tm.Cells[1].Population != 0 {
		t.Fatalf("expected unselected center to remain empty, got %.0f", tm.Cells[1].Population)
	}
	if tm.Cells[0].Population != 11 || tm.Cells[2].Population != 9 {
		t.Fatalf(
			"unexpected seeded distribution: got [%.0f %.0f], want [11 9]",
			tm.Cells[0].Population,
			tm.Cells[2].Population,
		)
	}
}

func TestPopulationModel_SeedPopulation_ClustersInitialPopulation(t *testing.T) {
	pm := simulation.NewPopulationModel(simulation.PopulationParameters{MaxCellCapacity: 10})
	tm := terrainMapWithFoodCapacities(10, 9, 8, 7, 6, 5, 4, 3, 2, 1)

	allocated := pm.SeedPopulation(tm, 40)
	if allocated != 40 {
		t.Fatalf("unexpected seeded total: got %d, want %d", allocated, 40)
	}

	occupied := make([]int, 0)
	for i, cell := range tm.Cells {
		if cell.Population > 0 {
			occupied = append(occupied, i)
		}
	}
	if len(occupied) != 2 {
		t.Fatalf("expected two initial population centers, got %d: %v", len(occupied), occupied)
	}
	if occupied[0] != 0 || occupied[1] != 2 {
		t.Fatalf("expected highest-capacity separated cells to seed population, got %v", occupied)
	}
}

func TestPopulationModel_SeedPopulation_ExpandsCentersToFitRequestedPopulation(t *testing.T) {
	pm := simulation.NewPopulationModel(simulation.PopulationParameters{MaxCellCapacity: 10})
	tm := terrainMapWithFoodCapacities(1, 1, 1, 1, 1, 1, 1, 1, 1, 1)

	allocated := pm.SeedPopulation(tm, 60)
	if allocated != 60 {
		t.Fatalf("unexpected seeded total: got %d, want %d", allocated, 60)
	}

	var total float64
	occupied := 0
	for _, cell := range tm.Cells {
		total += cell.Population
		if cell.Population > 0 {
			occupied++
		}
	}
	if total != 60 {
		t.Fatalf("expected exactly 60 people to be seeded, got %.0f", total)
	}
	if occupied < 3 {
		t.Fatalf("expected seeding to expand beyond the initial centers, got %d occupied cells", occupied)
	}
}

func TestPopulationModel_SeedPopulation_RespectsTotalCapacity(t *testing.T) {
	pm := simulation.NewPopulationModel(simulation.PopulationParameters{MaxCellCapacity: 1})
	tm := terrainMapWithFoodCapacities(9, 1)

	allocated := pm.SeedPopulation(tm, 15)
	if allocated != 10 {
		t.Fatalf("unexpected seeded total with cap: got %d, want %d", allocated, 10)
	}

	if tm.Cells[0].Population != 9 || tm.Cells[1].Population != 1 {
		t.Fatalf(
			"unexpected capped distribution: got [%.0f %.0f], want [9 1]",
			tm.Cells[0].Population,
			tm.Cells[1].Population,
		)
	}
}

func TestPopulationModel_SeedPopulation_SkipsNonPositiveFoodAndImpassableCells(t *testing.T) {
	pm := simulation.NewPopulationModel(simulation.PopulationParameters{MaxCellCapacity: 10})
	tm := terrainMapWithFoodCapacities(-1, 0, 2, 10)
	tm.Cells[3].Terrain = simtypes.TerrainTypeMountain

	allocated := pm.SeedPopulation(tm, 10)
	if allocated != 10 {
		t.Fatalf("unexpected seeded total: got %d, want %d", allocated, 10)
	}

	if tm.Cells[0].Population != 0 || tm.Cells[1].Population != 0 || tm.Cells[3].Population != 0 {
		t.Fatalf(
			"expected non-positive food and impassable cells to remain empty, got [%.0f %.0f %.0f]",
			tm.Cells[0].Population,
			tm.Cells[1].Population,
			tm.Cells[3].Population,
		)
	}
	if tm.Cells[2].Population != 10 {
		t.Fatalf("unexpected seeded population in positive-capacity cell: got %.0f, want 10", tm.Cells[2].Population)
	}
}

func TestPopulationModel_SeedPopulationBroadly_SpreadsAcrossViableCells(t *testing.T) {
	pm := simulation.NewPopulationModel(simulation.PopulationParameters{MaxCellCapacity: 10})
	tm := terrainMapWithFoodCapacities(1, 1, 1, 1, 1, 1)

	allocated := pm.SeedPopulationBroadly(tm, 6)
	if allocated != 6 {
		t.Fatalf("unexpected seeded total: got %d, want %d", allocated, 6)
	}

	occupied := 0
	for _, cell := range tm.Cells {
		if cell.Population > 0 {
			occupied++
		}
	}

	if occupied != 6 {
		t.Fatalf("expected broad seeding to occupy all viable cells, got %d occupied", occupied)
	}
}

func TestPopulationModel_AdvanceOneYear_GrowsBelowCarryingCapacity(t *testing.T) {
	pm := simulation.NewPopulationModel(simulation.PopulationParameters{
		GrowthRate:      0.025,
		MaxCellCapacity: 20,
	})
	tm := terrainMapWithFoodCapacities(1)
	tm.Cells[0].Population = 5

	pm.AdvanceOneYear(tm)

	if tm.Cells[0].Population <= 5 {
		t.Fatalf("expected population growth, got %.4f", tm.Cells[0].Population)
	}
	if pm.CurrentTick() != 1 {
		t.Fatalf("expected tick 1, got %d", pm.CurrentTick())
	}
}

func TestPopulationModel_AdvanceOneYear_AppliesStarvationAboveCapacity(t *testing.T) {
	pm := simulation.NewPopulationModel(simulation.PopulationParameters{
		GrowthRate:      0,
		StarvationRate:  0.5,
		MaxCellCapacity: 10,
	})
	tm := terrainMapWithFoodCapacities(1)
	tm.Cells[0].Population = 15

	tick := pm.AdvanceOneYear(tm)

	if tm.Cells[0].Population != 12.5 {
		t.Fatalf("expected starvation to reduce population to 12.5, got %.4f", tm.Cells[0].Population)
	}
	if tick.StarvationDeaths != 2.5 {
		t.Fatalf("expected starvation deaths 2.5, got %.4f", tick.StarvationDeaths)
	}
}

func TestPopulationModel_AdvanceOneYear_MigratesExcessPopulationToNeighbor(t *testing.T) {
	pm := simulation.NewPopulationModel(simulation.PopulationParameters{
		GrowthRate:      0,
		StarvationRate:  0,
		MigrationRate:   0.5,
		MaxCellCapacity: 10,
	})
	tm := terrainMapWithFoodCapacities(1, 1, 1)
	tm.Cells[0].Population = 20
	tm.Cells[1].Population = 0
	tm.Cells[2].Population = 0

	tick := pm.AdvanceOneYear(tm)

	if tm.Cells[0].Population != 15 {
		t.Fatalf("expected source population 15, got %.4f", tm.Cells[0].Population)
	}
	if tm.Cells[1].Population != 5 {
		t.Fatalf("expected neighbor population 5, got %.4f", tm.Cells[1].Population)
	}
	if tm.Cells[2].Population != 0 {
		t.Fatalf("expected non-neighbor population 0, got %.4f", tm.Cells[2].Population)
	}
	if tick.MigratedPopulation != 5 {
		t.Fatalf("expected migrated population 5, got %.4f", tick.MigratedPopulation)
	}
}

func TestPopulationModel_AdvanceOneYear_MigrationUsesSynchronousFlows(t *testing.T) {
	pm := simulation.NewPopulationModel(simulation.PopulationParameters{
		GrowthRate:      0,
		StarvationRate:  0,
		MigrationRate:   0.5,
		MaxCellCapacity: 10,
	})
	tm := terrainMapWithFoodCapacities(1, 1, 1)
	tm.Cells[0].Population = 20
	tm.Cells[1].Population = 0
	tm.Cells[2].Population = 0

	pm.AdvanceOneYear(tm)

	if tm.Cells[1].Population != 5 {
		t.Fatalf("expected synchronous migration to leave neighbor at 5, got %.4f", tm.Cells[1].Population)
	}
}

func terrainMapWithFoodCapacities(foodCapacities ...float64) *simtypes.TerrainMap {
	tm := simtypes.NewTerrainMap(len(foodCapacities), 1)
	for i, capacity := range foodCapacities {
		tm.Cells[i].Location = simtypes.Position{X: i, Y: 0}
		tm.Cells[i].Terrain = simtypes.TerrainTypePlains
		tm.Cells[i].FoodCapacity = capacity
		tm.Cells[i].Population = 0
	}
	return tm
}
