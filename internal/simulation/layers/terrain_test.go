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
	elevation.Set(1, 1, layers.WaterElevationThreshold-0.01)

	if err := classifier.Classify(elevation, moisture, fertility); err != nil {
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
	elevation.Set(1, 1, layers.MountainElevationThreshold+0.01)

	if err := classifier.Classify(elevation, moisture, fertility); err != nil {
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

func TestTerrainClassifier_Classify_ModerateHighMoistureAndFertilityCanProduceForest(t *testing.T) {
	tm := terrainMapWithLocations(3, 3)
	classifier := layers.NewTerrainClassifier(tm)
	elevation := uniformLayer(3, 3, 0.6)
	moisture := uniformLayer(3, 3, 0.9)
	fertility := uniformLayer(3, 3, 0.9)

	if err := classifier.Classify(elevation, moisture, fertility); err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	center := tm.GetCell(1, 1)
	if center.Terrain != simtypes.TerrainTypeForest {
		t.Fatalf("expected center terrain to be Forest, got %s", center.Terrain.String())
	}
}

func TestTerrainClassifier_Classify_ModerateLowerMoistureAndFertilityCanProducePlains(t *testing.T) {
	tm := terrainMapWithLocations(3, 3)
	classifier := layers.NewTerrainClassifier(tm)
	elevation := uniformLayer(3, 3, 0.55)
	moisture := uniformLayer(3, 3, 0.3)
	fertility := uniformLayer(3, 3, 0.2)

	if err := classifier.Classify(elevation, moisture, fertility); err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	center := tm.GetCell(1, 1)
	if center.Terrain != simtypes.TerrainTypePlains {
		t.Fatalf("expected center terrain to be Plains, got %s", center.Terrain.String())
	}
}

func TestTerrainClassifier_Classify_NeighborInfluenceAffectsButDoesNotForceOutcome(t *testing.T) {
	plainsCenter := classifyCenterWithNeighborProfile(t, 0.3, 0.2)
	if plainsCenter != simtypes.TerrainTypePlains {
		t.Fatalf("expected plains-biased neighborhood to keep center Plains, got %s", plainsCenter.String())
	}

	forestCenter := classifyCenterWithNeighborProfile(t, 0.9, 0.9)
	if forestCenter != simtypes.TerrainTypeForest {
		t.Fatalf("expected forest-biased neighborhood to shift center to Forest, got %s", forestCenter.String())
	}
}

func TestTerrainClassifier_Classify_ElevationThresholdMatrix(t *testing.T) {
	tests := []struct {
		name        string
		elevation   float64
		wantTerrain simtypes.TerrainType
	}{
		{
			name:        "below water threshold is water",
			elevation:   layers.WaterElevationThreshold - 0.001,
			wantTerrain: simtypes.TerrainTypeWater,
		},
		{
			name:        "exact water threshold is water",
			elevation:   layers.WaterElevationThreshold,
			wantTerrain: simtypes.TerrainTypeWater,
		},
		{
			name:        "just above water threshold defaults to water under neutral conditions",
			elevation:   layers.WaterElevationThreshold + 0.001,
			wantTerrain: simtypes.TerrainTypeWater,
		},
		{
			name:        "just below mountain threshold defaults to mountain under neutral conditions",
			elevation:   layers.MountainElevationThreshold - 0.001,
			wantTerrain: simtypes.TerrainTypeMountain,
		},
		{
			name:        "exact mountain threshold is mountain",
			elevation:   layers.MountainElevationThreshold,
			wantTerrain: simtypes.TerrainTypeMountain,
		},
		{
			name:        "above mountain threshold is mountain",
			elevation:   layers.MountainElevationThreshold + 0.001,
			wantTerrain: simtypes.TerrainTypeMountain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := terrainMapWithLocations(3, 3)
			classifier := layers.NewTerrainClassifier(tm)

			elevation := uniformLayer(3, 3, 0.5)
			moisture := uniformLayer(3, 3, 0.5)
			fertility := uniformLayer(3, 3, 0.5)
			elevation.Set(1, 1, tt.elevation)

			if err := classifier.Classify(elevation, moisture, fertility); err != nil {
				t.Fatalf("Classify returned error: %v", err)
			}

			center := tm.GetCell(1, 1)
			if center == nil {
				t.Fatal("expected center cell to exist")
			}

			if center.Terrain != tt.wantTerrain {
				t.Fatalf("expected center terrain %s, got %s", tt.wantTerrain.String(), center.Terrain.String())
			}
		})
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

func classifyCenterWithNeighborProfile(t *testing.T, neighborMoisture, neighborFertility float64) simtypes.TerrainType {
	t.Helper()

	tm := terrainMapWithLocations(3, 3)
	classifier := layers.NewTerrainClassifier(tm)

	elevation := uniformLayer(3, 3, 0.55)
	moisture := uniformLayer(3, 3, neighborMoisture)
	fertility := uniformLayer(3, 3, neighborFertility)

	// Keep the center near a plains/forest tie so neighborhood context can influence it.
	elevation.Set(1, 1, 0.5)
	moisture.Set(1, 1, 0.45)
	fertility.Set(1, 1, 0.45)

	if err := classifier.Classify(elevation, moisture, fertility); err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	return tm.GetCell(1, 1).Terrain
}
