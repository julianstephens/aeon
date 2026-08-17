package layers_test

import (
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/layers"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

func TestPipeline_Run_AppliesScalarLayersAndAssignsTerrain(t *testing.T) {
	worldSeed := seedFromString("pipeline-applies-layers")
	pipeline := layers.NewPipeline(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight, rng.NewRNG(worldSeed))

	tm, err := pipeline.Run(worldSeed)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if tm == nil {
		t.Fatal("expected pipeline to return terrain map")
	}
	if !tm.IsInitialized() {
		t.Fatal("expected returned terrain map to be initialized")
	}

	hasNonDefaultScalar := false
	for i, cell := range tm.Cells {
		if cell.Elevation < 0 || cell.Elevation > 1 {
			t.Fatalf("elevation out of [0,1] at index %d: %f", i, cell.Elevation)
		}
		if cell.Moisture < 0 || cell.Moisture > 1 {
			t.Fatalf("moisture out of [0,1] at index %d: %f", i, cell.Moisture)
		}
		if cell.Fertility < 0 || cell.Fertility > 1 {
			t.Fatalf("fertility out of [0,1] at index %d: %f", i, cell.Fertility)
		}

		if cell.Elevation != 0 || cell.Moisture != 0 || cell.Fertility != 0 {
			hasNonDefaultScalar = true
		}

		if cell.Terrain != simtypes.TerrainTypeWater &&
			cell.Terrain != simtypes.TerrainTypePlains &&
			cell.Terrain != simtypes.TerrainTypeForest &&
			cell.Terrain != simtypes.TerrainTypeMountain {
			t.Fatalf("invalid terrain type at index %d: %v", i, cell.Terrain)
		}
	}

	if !hasNonDefaultScalar {
		t.Fatal("expected scalar layers to be applied to terrain map")
	}
}

func TestPipeline_Run_IsDeterministicForSameSeed(t *testing.T) {
	worldSeed := seedFromString("pipeline-determinism")
	pipeline := layers.NewPipeline(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight, rng.NewRNG(worldSeed))

	first, err := pipeline.Run(worldSeed)
	if err != nil {
		t.Fatalf("first Run returned error: %v", err)
	}
	firstSnapshot := cloneTerrainMap(*first)

	second, err := pipeline.Run(worldSeed)
	if err != nil {
		t.Fatalf("second Run returned error: %v", err)
	}

	assertTerrainMapsEqual(t, firstSnapshot, *second)
}

func cloneTerrainMap(src simtypes.TerrainMap) simtypes.TerrainMap {
	cells := make([]simtypes.TerrainCell, len(src.Cells))
	copy(cells, src.Cells)

	return simtypes.TerrainMap{
		Width:  src.Width,
		Height: src.Height,
		Cells:  cells,
	}
}

func assertTerrainMapsEqual(t *testing.T, a, b simtypes.TerrainMap) {
	t.Helper()

	if a.Width != b.Width || a.Height != b.Height {
		t.Fatalf("map shape mismatch: (%d,%d) vs (%d,%d)", a.Width, a.Height, b.Width, b.Height)
	}

	for i := range a.Cells {
		cellA := a.Cells[i]
		cellB := b.Cells[i]

		if cellA.Elevation != cellB.Elevation {
			t.Fatalf("elevation mismatch at index %d: %f != %f", i, cellA.Elevation, cellB.Elevation)
		}
		if cellA.Moisture != cellB.Moisture {
			t.Fatalf("moisture mismatch at index %d: %f != %f", i, cellA.Moisture, cellB.Moisture)
		}
		if cellA.Fertility != cellB.Fertility {
			t.Fatalf("fertility mismatch at index %d: %f != %f", i, cellA.Fertility, cellB.Fertility)
		}
		if cellA.Terrain != cellB.Terrain {
			t.Fatalf("terrain mismatch at index %d: %s != %s", i, cellA.Terrain.String(), cellB.Terrain.String())
		}
	}
}
