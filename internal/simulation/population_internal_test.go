package simulation

import (
	"math"
	"reflect"
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
)

func TestPopulationModel_GrowPopulation_GrowsBelowCapacity(t *testing.T) {
	pm := NewPopulationModel(PopulationParameters{GrowthRate: 0.2, MaxCellCapacity: 1})
	tm := terrainMapWithPopulation(10, 4)

	before := tm.Cells[0].Population
	pm.growPopulation(tm)
	after := tm.Cells[0].Population

	if after <= before {
		t.Fatalf("expected population to grow below capacity, got %.2f -> %.2f", before, after)
	}
	if after >= 10 {
		t.Fatalf("expected population to remain below capacity, got %.2f", after)
	}
}

func TestPopulationModel_GrowPopulation_ApproachesZeroNearCapacity(t *testing.T) {
	pm := NewPopulationModel(PopulationParameters{GrowthRate: 0.2, MaxCellCapacity: 1})
	low := terrainMapWithPopulation(10, 5)
	near := terrainMapWithPopulation(10, 9)

	lowBefore := low.Cells[0].Population
	nearBefore := near.Cells[0].Population

	pm.growPopulation(low)
	pm.growPopulation(near)

	lowGrowth := low.Cells[0].Population - lowBefore
	nearGrowth := near.Cells[0].Population - nearBefore

	if lowGrowth <= 0 || nearGrowth <= 0 {
		t.Fatalf("expected both populations to increase: low %.2f, near %.2f", lowGrowth, nearGrowth)
	}
	if nearGrowth >= lowGrowth {
		t.Fatalf(
			"expected growth near capacity to be smaller than growth at lower density: low=%.4f near=%.4f",
			lowGrowth,
			nearGrowth,
		)
	}
	if math.Abs(nearGrowth) < 0.000001 {
		t.Fatalf("expected growth to approach zero near capacity, got %.6f", nearGrowth)
	}
}

func TestPopulationModel_GrowPopulation_DeclinesAboveCapacity(t *testing.T) {
	pm := NewPopulationModel(PopulationParameters{GrowthRate: 0.2, MaxCellCapacity: 1})
	tm := terrainMapWithPopulation(10, 12)

	before := tm.Cells[0].Population
	pm.growPopulation(tm)
	after := tm.Cells[0].Population

	if after >= before {
		t.Fatalf("expected population to decline above capacity, got %.2f -> %.2f", before, after)
	}
	if after < 0 {
		t.Fatalf("expected population to remain non-negative above capacity, got %.2f", after)
	}
}

func TestPopulationModel_GrowPopulation_ZeroFoodCellsCannotSustainPopulation(t *testing.T) {
	pm := NewPopulationModel(PopulationParameters{GrowthRate: 0.2, MaxCellCapacity: 1})
	tm := terrainMapWithPopulation(0, 6)

	pm.growPopulation(tm)
	if tm.Cells[0].Population != 0 {
		t.Fatalf("expected zero-food cell to reset population to zero, got %.2f", tm.Cells[0].Population)
	}
}

func TestPopulationModel_GrowPopulation_NeverBecomesNegative(t *testing.T) {
	pm := NewPopulationModel(PopulationParameters{GrowthRate: 0.2, MaxCellCapacity: 2})
	tm := terrainMapWithPopulation(10, -8)

	pm.growPopulation(tm)
	if tm.Cells[0].Population < 0 {
		t.Fatalf("expected population clamp to prevent negative values, got %.2f", tm.Cells[0].Population)
	}
}

func TestPopulationModel_GrowPopulation_RemainsDeterministic(t *testing.T) {
	params := PopulationParameters{GrowthRate: 0.25, MaxCellCapacity: 1}
	m1 := terrainMapWithPopulationSet([]float64{10, 12, 8}, []float64{4, 7, 3})
	m2 := terrainMapWithPopulationSet([]float64{10, 12, 8}, []float64{4, 7, 3})

	pm1 := NewPopulationModel(params)
	pm2 := NewPopulationModel(params)

	for i := 0; i < 20; i++ {
		pm1.AdvanceOneYear(m1)
		pm2.AdvanceOneYear(m2)
	}

	if !reflect.DeepEqual(m1.Cells, m2.Cells) {
		t.Fatalf("expected deterministic population evolution, got %#v and %#v", m1.Cells, m2.Cells)
	}
}

func terrainMapWithPopulation(foodCapacity, population float64) *simtypes.TerrainMap {
	return terrainMapWithPopulationSet([]float64{foodCapacity}, []float64{population})
}

func terrainMapWithPopulationSet(foodCapacities, populations []float64) *simtypes.TerrainMap {
	tm := simtypes.NewTerrainMap(len(foodCapacities), 1)
	for i, foodCapacity := range foodCapacities {
		tm.Cells[i].Location = simtypes.Position{X: i, Y: 0}
		tm.Cells[i].Terrain = simtypes.TerrainTypePlains
		tm.Cells[i].FoodCapacity = foodCapacity
		tm.Cells[i].Population = populations[i]
	}
	return tm
}
