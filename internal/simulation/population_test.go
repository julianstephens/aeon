package simulation_test

import (
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation"
)

func TestPopulationModel_SeedPopulation_WeightedByFoodCapacity(t *testing.T) {
	pm := simulation.NewPopulationModel(simulation.PopulationParameters{MaxCellCapacity: 100})
	tm := terrainMapWithFoodCapacities(1, 2, 3)

	allocated := pm.SeedPopulation(tm, 60)
	if allocated != 60 {
		t.Fatalf("unexpected seeded total: got %d, want %d", allocated, 60)
	}

	want := []float64{10, 20, 30}
	for i := range want {
		if tm.Cells[i].Population != want[i] {
			t.Fatalf("unexpected cell %d population: got %.0f, want %.0f", i, tm.Cells[i].Population, want[i])
		}
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

func terrainMapWithFoodCapacities(foodCapacities ...float64) *simtypes.TerrainMap {
	tm := simtypes.NewTerrainMap(len(foodCapacities), 1)
	for i, capacity := range foodCapacities {
		tm.Cells[i].Location = simtypes.Position{X: i, Y: 0}
		tm.Cells[i].FoodCapacity = capacity
		tm.Cells[i].Population = 99
	}
	return tm
}
