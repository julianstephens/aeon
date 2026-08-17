package layers_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/layers"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

var diagnosticSeeds = []uint64{42, 43, 44, 1337, 9001, 123456}

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

func TestPipeline_GenerateWorld_IsDeterministicForSameSeed(t *testing.T) {
	seed := seedFromString("pipeline-retry-determinism")
	pipelineA := layers.NewPipeline(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight, rng.NewRNG(seed))
	pipelineB := layers.NewPipeline(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight, rng.NewRNG(seed))

	first, err := pipelineA.GenerateWorld(seed)
	if err != nil {
		t.Fatalf("first GenerateWorld returned error: %v", err)
	}

	second, err := pipelineB.GenerateWorld(seed)
	if err != nil {
		t.Fatalf("second GenerateWorld returned error: %v", err)
	}

	assertTerrainMapsEqual(t, *first, *second)
}

func TestPipeline_GenerateWorld_DiagnosticSeedCorpusProducesViableWorlds(t *testing.T) {
	for _, seedValue := range diagnosticSeeds {
		t.Run(fmt.Sprintf("seed_%d", seedValue), func(t *testing.T) {
			seed := sha256.Sum256(fmt.Appendf(nil, "%d", seedValue))
			pipeline := layers.NewPipeline(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight, rng.NewRNG(seed))

			tm, err := pipeline.GenerateWorld(seed)
			if err != nil {
				t.Fatalf("GenerateWorld returned error: %v", err)
			}

			diagnostics := layers.ComputeWorldDiagnostics(*tm)
			if err := layers.ValidateWorld(diagnostics, layers.DefaultViabilityRules()); err != nil {
				t.Fatalf("generated world is not viable: %v", err)
			}

			terrainKinds := map[simtypes.TerrainType]struct{}{}
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
				terrainKinds[cell.Terrain] = struct{}{}
			}

			if len(terrainKinds) < 2 {
				t.Fatalf("expected at least two terrain types, got %d", len(terrainKinds))
			}

			if diagnostics.LargestPassableLandRegion <= 0 {
				t.Fatal("expected at least one passable land region")
			}
		})
	}
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
