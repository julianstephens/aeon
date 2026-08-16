package layers_test

import (
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/layers"
)

func TestTerrainClassifier_Classify_SetsWaterFromLowElevation(t *testing.T) {
	tm := terrainMapWithLocations(3, 3)
	classifier := layers.NewTerrainClassifier(tm)

	elevation := uniformLayer(3, 3, 0.5)
	moisture := uniformLayer(3, 3, 0.5)
	fertility := uniformLayer(3, 3, 0.5)
	terrain := uniformLayer(3, 3, 1)
	elevation.Set(1, 1, layers.WaterThreshold-0.01)

	if err := classifier.Classify(elevation, moisture, fertility, terrain); err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	center := tm.GetCell(1, 1)
	if center == nil {
		t.Fatal("expected center cell to exist")
	}
	if center.Terrain != simtypes.TerrainTypeWater {
		t.Fatalf("expected center terrain to be Water, got %s", center.Terrain.String())
	}
}

func TestTerrainClassifier_Classify_SetsMountainFromHighElevation(t *testing.T) {
	tm := terrainMapWithLocations(3, 3)
	classifier := layers.NewTerrainClassifier(tm)

	elevation := uniformLayer(3, 3, 0.5)
	moisture := uniformLayer(3, 3, 0.5)
	fertility := uniformLayer(3, 3, 0.5)
	terrain := uniformLayer(3, 3, 1)
	elevation.Set(1, 1, layers.MountainThreshold+0.01)

	if err := classifier.Classify(elevation, moisture, fertility, terrain); err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	center := tm.GetCell(1, 1)
	if center == nil {
		t.Fatal("expected center cell to exist")
	}
	if center.Terrain != simtypes.TerrainTypeMountain {
		t.Fatalf("expected center terrain to be Mountain, got %s", center.Terrain.String())
	}
}

func TestTerrainClassifier_Classify_ConvertsIsolatedForestToPlains(t *testing.T) {
	tm := terrainMapWithLocations(3, 3)
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			cell := tm.GetCell(x, y)
			cell.Terrain = simtypes.TerrainTypePlains
			tm.SetCell(x, y, *cell)
		}
	}
	center := tm.GetCell(1, 1)
	center.Terrain = simtypes.TerrainTypeForest
	tm.SetCell(1, 1, *center)

	classifier := layers.NewTerrainClassifier(tm)
	elevation := uniformLayer(3, 3, 0.5)
	moisture := uniformLayer(3, 3, 0.5)
	fertility := uniformLayer(3, 3, 0.5)
	terrain := uniformLayer(3, 3, 1)

	if err := classifier.Classify(elevation, moisture, fertility, terrain); err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	updated := tm.GetCell(1, 1)
	if updated.Terrain != simtypes.TerrainTypePlains {
		t.Fatalf("expected isolated forest to become Plains, got %s", updated.Terrain.String())
	}
}

func TestTerrainClassifier_Classify_ConvertsIsolatedPlainsToForest(t *testing.T) {
	tm := terrainMapWithLocations(3, 3)
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			cell := tm.GetCell(x, y)
			cell.Terrain = simtypes.TerrainTypeForest
			tm.SetCell(x, y, *cell)
		}
	}
	center := tm.GetCell(1, 1)
	center.Terrain = simtypes.TerrainTypePlains
	tm.SetCell(1, 1, *center)

	classifier := layers.NewTerrainClassifier(tm)
	elevation := uniformLayer(3, 3, 0.5)
	moisture := uniformLayer(3, 3, 0.5)
	fertility := uniformLayer(3, 3, 0.5)
	terrain := uniformLayer(3, 3, 1)

	if err := classifier.Classify(elevation, moisture, fertility, terrain); err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	updated := tm.GetCell(1, 1)
	if updated.Terrain != simtypes.TerrainTypeForest {
		t.Fatalf("expected isolated plains to become Forest, got %s", updated.Terrain.String())
	}
}

func TestTerrainClassifier_Classify_ConvertsIsolatedWaterToDominantNeighborType(t *testing.T) {
	tm := terrainMapWithLocations(3, 3)
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			cell := tm.GetCell(x, y)
			cell.Terrain = simtypes.TerrainTypePlains
			tm.SetCell(x, y, *cell)
		}
	}

	// Create a mixed neighborhood with forest as the dominant non-water type.
	forestNeighbors := [][2]int{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {2, 1}}
	for _, pos := range forestNeighbors {
		cell := tm.GetCell(pos[0], pos[1])
		cell.Terrain = simtypes.TerrainTypeForest
		tm.SetCell(pos[0], pos[1], *cell)
	}

	center := tm.GetCell(1, 1)
	center.Terrain = simtypes.TerrainTypeWater
	tm.SetCell(1, 1, *center)

	classifier := layers.NewTerrainClassifier(tm)
	elevation := uniformLayer(3, 3, 0.5)
	moisture := uniformLayer(3, 3, 0.5)
	fertility := uniformLayer(3, 3, 0.5)
	terrain := uniformLayer(3, 3, 1)

	if err := classifier.Classify(elevation, moisture, fertility, terrain); err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	updated := tm.GetCell(1, 1)
	if updated.Terrain != simtypes.TerrainTypeForest {
		t.Fatalf("expected isolated water to become Forest, got %s", updated.Terrain.String())
	}
}

func uniformLayer(width, height int, value float64) *simtypes.Layer {
	l := simtypes.NewLayer(width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			l.Set(x, y, value)
		}
	}
	return l
}

func terrainMapWithLocations(width, height int) *simtypes.TerrainMap {
	tm := simtypes.NewTerrainMap(width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := tm.GetCell(x, y)
			cell.Location = simtypes.Position{X: x, Y: y}
			tm.SetCell(x, y, *cell)
		}
	}
	return tm
}
