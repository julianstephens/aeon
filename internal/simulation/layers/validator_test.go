package layers_test

import (
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/layers"
)

func TestValidateWorld_AllowsHighForestComposition(t *testing.T) {
	tm := filledTerrainMap(20, 20, simtypes.TerrainTypeForest)
	overwriteRect(tm, 0, 0, 5, 20, simtypes.TerrainTypeWater)

	diagnostics := layers.ComputeWorldDiagnostics(*tm)
	if err := layers.ValidateWorld(diagnostics, layers.DefaultViabilityRules()); err != nil {
		t.Fatalf("expected high-forest world to be viable, got error: %v", err)
	}
}

func TestValidateWorld_AllowsLowForestComposition(t *testing.T) {
	tm := filledTerrainMap(20, 20, simtypes.TerrainTypePlains)
	overwriteRect(tm, 0, 0, 1, 20, simtypes.TerrainTypeForest)

	diagnostics := layers.ComputeWorldDiagnostics(*tm)
	if err := layers.ValidateWorld(diagnostics, layers.DefaultViabilityRules()); err != nil {
		t.Fatalf("expected low-forest world to be viable, got error: %v", err)
	}
}

func TestValidateWorld_RejectsExtremeWater(t *testing.T) {
	tm := filledTerrainMap(20, 20, simtypes.TerrainTypeWater)
	overwriteRect(tm, 0, 0, 1, 20, simtypes.TerrainTypePlains)

	diagnostics := layers.ComputeWorldDiagnostics(*tm)
	if err := layers.ValidateWorld(diagnostics, layers.DefaultViabilityRules()); err == nil {
		t.Fatal("expected extreme-water world to be invalid")
	}
}

func TestValidateWorld_RejectsFragmentedLandmass(t *testing.T) {
	tm := filledTerrainMap(20, 20, simtypes.TerrainTypeWater)

	// Place isolated passable cells so total land can exceed minimum while
	// largest contiguous passable region remains very small.
	for y := 0; y < tm.Height; y += 2 {
		for x := 0; x < tm.Width; x += 2 {
			terrain := simtypes.TerrainTypePlains
			if (x+y)%4 == 0 {
				terrain = simtypes.TerrainTypeForest
			}
			_ = tm.SetTerrainType(x, y, terrain)
		}
	}

	diagnostics := layers.ComputeWorldDiagnostics(*tm)
	err := layers.ValidateWorld(diagnostics, layers.DefaultViabilityRules())
	if err == nil {
		t.Fatal("expected fragmented world to be invalid")
	}

	validationErr, ok := err.(*layers.WorldValidationError)
	if !ok {
		t.Fatalf("expected WorldValidationError, got %T", err)
	}

	foundLargestRegionFailure := false
	for _, reason := range validationErr.Reasons {
		if len(reason) >= len("largest land region") && reason[:len("largest land region")] == "largest land region" {
			foundLargestRegionFailure = true
			break
		}
	}

	if !foundLargestRegionFailure {
		t.Fatalf("expected largest region failure in reasons: %v", validationErr.Reasons)
	}
}

func filledTerrainMap(width, height int, terrain simtypes.TerrainType) *simtypes.TerrainMap {
	tm := simtypes.NewTerrainMap(width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			_ = tm.SetTerrainType(x, y, terrain)
		}
	}
	return tm
}

func overwriteRect(tm *simtypes.TerrainMap, startX, startY, rectWidth, rectHeight int, terrain simtypes.TerrainType) {
	for y := startY; y < startY+rectHeight; y++ {
		for x := startX; x < startX+rectWidth; x++ {
			_ = tm.SetTerrainType(x, y, terrain)
		}
	}
}
