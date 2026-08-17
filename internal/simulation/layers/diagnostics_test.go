package layers_test

import (
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/layers"
)

func TestComputeWorldDiagnostics_ComputesDistributionAndRegions(t *testing.T) {
	tm := terrainMapFromRows([][]simtypes.TerrainType{
		{simtypes.TerrainTypePlains, simtypes.TerrainTypePlains, simtypes.TerrainTypeWater, simtypes.TerrainTypeWater},
		{
			simtypes.TerrainTypePlains,
			simtypes.TerrainTypeForest,
			simtypes.TerrainTypeWater,
			simtypes.TerrainTypeMountain,
		},
		{
			simtypes.TerrainTypeForest,
			simtypes.TerrainTypeForest,
			simtypes.TerrainTypeMountain,
			simtypes.TerrainTypeMountain,
		},
		{
			simtypes.TerrainTypePlains,
			simtypes.TerrainTypeWater,
			simtypes.TerrainTypeMountain,
			simtypes.TerrainTypeMountain,
		},
	})

	diagnostics := layers.ComputeWorldDiagnostics(*tm)

	if got, want := diagnostics.TerrainCounts[simtypes.TerrainTypePlains], 4; got != want {
		t.Fatalf("plains count mismatch: got %d want %d", got, want)
	}
	if got, want := diagnostics.TerrainCounts[simtypes.TerrainTypeForest], 3; got != want {
		t.Fatalf("forest count mismatch: got %d want %d", got, want)
	}
	if got, want := diagnostics.TerrainCounts[simtypes.TerrainTypeWater], 4; got != want {
		t.Fatalf("water count mismatch: got %d want %d", got, want)
	}
	if got, want := diagnostics.TerrainCounts[simtypes.TerrainTypeMountain], 5; got != want {
		t.Fatalf("mountain count mismatch: got %d want %d", got, want)
	}

	if got, want := diagnostics.RegionCounts[simtypes.TerrainTypePlains], 2; got != want {
		t.Fatalf("plains region count mismatch: got %d want %d", got, want)
	}
	if got, want := diagnostics.LargestRegion[simtypes.TerrainTypeMountain], 5; got != want {
		t.Fatalf("largest mountain region mismatch: got %d want %d", got, want)
	}
	if got, want := diagnostics.LargestPassableLandRegion, 7; got != want {
		t.Fatalf("largest passable region mismatch: got %d want %d", got, want)
	}

	if got, want := diagnostics.PassableLandPercentage, 7.0/16.0; got != want {
		t.Fatalf("passable percentage mismatch: got %f want %f", got, want)
	}
}

func terrainMapFromRows(rows [][]simtypes.TerrainType) *simtypes.TerrainMap {
	height := len(rows)
	width := 0
	if height > 0 {
		width = len(rows[0])
	}

	tm := simtypes.NewTerrainMap(width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			_ = tm.SetTerrainType(x, y, rows[y][x])
		}
	}
	return tm
}
