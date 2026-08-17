package main

import (
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
)

func TestFoodCapacityStatsByTerrain(t *testing.T) {
	tm := simtypes.NewTerrainMap(3, 2)

	cells := []struct {
		x, y    int
		terrain simtypes.TerrainType
		value   float64
	}{
		{0, 0, simtypes.TerrainTypePlains, 0.2},
		{1, 0, simtypes.TerrainTypePlains, 0.8},
		{2, 0, simtypes.TerrainTypeForest, 0.4},
		{0, 1, simtypes.TerrainTypeForest, 0.6},
		{1, 1, simtypes.TerrainTypeWater, 0.0},
		{2, 1, simtypes.TerrainTypeMountain, 0.3},
	}

	for _, cell := range cells {
		c := tm.GetCell(cell.x, cell.y)
		c.Terrain = cell.terrain
		c.FoodCapacity = cell.value
		tm.SetCell(cell.x, cell.y, *c)
	}

	stats := foodCapacityStatsByTerrain(*tm)

	if got := stats[simtypes.TerrainTypePlains].Mean; got != 0.5 {
		t.Fatalf("plains mean = %v, want 0.5", got)
	}
	if got := stats[simtypes.TerrainTypePlains].Max; got != 0.8 {
		t.Fatalf("plains max = %v, want 0.8", got)
	}
	if got := stats[simtypes.TerrainTypeForest].Mean; got != 0.5 {
		t.Fatalf("forest mean = %v, want 0.5", got)
	}
	if got := stats[simtypes.TerrainTypeWater].Mean; got != 0.0 {
		t.Fatalf("water mean = %v, want 0.0", got)
	}
	if got := stats[simtypes.TerrainTypeWater].Max; got != 0.0 {
		t.Fatalf("water max = %v, want 0.0", got)
	}
}
