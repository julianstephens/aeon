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

func TestPopulationModel_IsolatedSingleCell_ApproachesLocalCapacity(t *testing.T) {
	pm := NewPopulationModel(PopulationParameters{
		GrowthRate:      0.2,
		StarvationRate:  0,
		MigrationRate:   0,
		MaxCellCapacity: 10,
	})
	tm := terrainMapWithPopulation(100, 500)

	for i := 0; i < 200; i++ {
		pm.AdvanceOneYear(tm)
	}

	finalPopulation := tm.Cells[0].Population
	if math.Abs(finalPopulation-1000) > 0.1 {
		t.Fatalf("expected isolated cell population to approach 1000, got %.4f", finalPopulation)
	}
}

func TestPopulationModel_IsolatedIndependentCells_ApproachSumOfLocalCapacities(t *testing.T) {
	pm := NewPopulationModel(PopulationParameters{
		GrowthRate:      0.2,
		StarvationRate:  0,
		MigrationRate:   0,
		MaxCellCapacity: 10,
	})
	tm := terrainMapWithPopulationSet(
		[]float64{100, 50, 25},
		[]float64{500, 250, 125},
	)

	for i := 0; i < 220; i++ {
		pm.AdvanceOneYear(tm)
	}

	total := tm.Cells[0].Population + tm.Cells[1].Population + tm.Cells[2].Population
	if math.Abs(total-1750) > 0.2 {
		t.Fatalf("expected total population to approach 1750, got %.4f", total)
	}
}

func TestPopulationModel_MigrationSummaryTracksFlowStats(t *testing.T) {
	pm := NewPopulationModel(PopulationParameters{
		GrowthRate:      0,
		StarvationRate:  0,
		MigrationRate:   0.5,
		MaxCellCapacity: 10,
	})
	terrain := simtypes.NewTerrainMap(3, 1)
	for i := range terrain.Cells {
		terrain.Cells[i].Location = simtypes.Position{X: i, Y: 0}
		terrain.Cells[i].Terrain = simtypes.TerrainTypePlains
		terrain.Cells[i].FoodCapacity = 1
	}
	terrain.Cells[0].Population = 20
	terrain.Cells[1].Population = 0
	terrain.Cells[2].Population = 0

	pm.AdvanceOneYear(terrain)
	summary := pm.MigrationSummary()

	if summary.Moved != 5 {
		t.Fatalf("expected 5 migrated people, got %.0f", summary.Moved)
	}
	if summary.SourceCells != 1 {
		t.Fatalf("expected 1 source cell, got %d", summary.SourceCells)
	}
	if summary.DestinationCells != 1 {
		t.Fatalf("expected 1 destination cell, got %d", summary.DestinationCells)
	}
	if summary.AverageDistance != 1.0 {
		t.Fatalf("expected average distance 1.0, got %.1f", summary.AverageDistance)
	}
	if summary.MaxDistance != 1 {
		t.Fatalf("expected max distance 1, got %d", summary.MaxDistance)
	}
}

func TestPopulationModel_PopulationPressureSummary_ClassifiesCellsByUtilization(t *testing.T) {
	pm := NewPopulationModel(PopulationParameters{MaxCellCapacity: 10})
	terrain := simtypes.NewTerrainMap(5, 1)
	for i := range terrain.Cells {
		terrain.Cells[i].Location = simtypes.Position{X: i, Y: 0}
		terrain.Cells[i].Terrain = simtypes.TerrainTypePlains
		terrain.Cells[i].FoodCapacity = 1
	}
	terrain.Cells[0].Population = 2  // 20%
	terrain.Cells[1].Population = 3  // 30%
	terrain.Cells[2].Population = 6  // 60%
	terrain.Cells[3].Population = 8  // 80%
	terrain.Cells[4].Population = 15 // 150%

	summary := pm.PopulationPressureSummary(terrain)
	if summary.Under25 != 1 || summary.Range25to50 != 1 || summary.Range50to75 != 1 || summary.Range75to100 != 1 ||
		summary.Over100 != 1 {
		t.Fatalf("unexpected pressure buckets: %+v", summary)
	}
}

func TestTerrainMap_CreatesCellCoordinatesForPopulationModel(t *testing.T) {
	tm := simtypes.NewTerrainMap(3, 2)
	if tm.Cells[0].Location != (simtypes.Position{X: 0, Y: 0}) {
		t.Fatalf("expected top-left cell location to be (0,0), got %#v", tm.Cells[0].Location)
	}
	if tm.Cells[1].Location != (simtypes.Position{X: 1, Y: 0}) {
		t.Fatalf("expected (1,0) cell coordinates, got %#v", tm.Cells[1].Location)
	}
	if tm.Cells[3].Location != (simtypes.Position{X: 0, Y: 1}) {
		t.Fatalf("expected (0,1) cell coordinates, got %#v", tm.Cells[3].Location)
	}
	if tm.Cells[5].Location != (simtypes.Position{X: 2, Y: 1}) {
		t.Fatalf("expected bottom-right cell location to be (2,1), got %#v", tm.Cells[5].Location)
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
