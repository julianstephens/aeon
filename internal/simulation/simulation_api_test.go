package simulation

import "testing"

func TestNewSimulation_IsDeterministicForSameSeed(t *testing.T) {
	cfg := DefaultGUIConfig()
	cfg.Seed = "gui-seed-42"
	cfg.InitialPopulation = 1000

	first, err := NewSimulation(cfg.Seed, PopulationParameters{
		GrowthRate:      cfg.GrowthRate,
		StarvationRate:  cfg.StarvationRate,
		MigrationRate:   cfg.MigrationRate,
		MaxCellCapacity: cfg.MaxCellCapacity,
	})
	if err != nil {
		t.Fatalf("NewSimulation returned error: %v", err)
	}
	second, err := NewSimulation(cfg.Seed, PopulationParameters{
		GrowthRate:      cfg.GrowthRate,
		StarvationRate:  cfg.StarvationRate,
		MigrationRate:   cfg.MigrationRate,
		MaxCellCapacity: cfg.MaxCellCapacity,
	})
	if err != nil {
		t.Fatalf("second NewSimulation returned error: %v", err)
	}

	if first.Snapshot().Year != second.Snapshot().Year {
		t.Fatalf("year mismatch: %d vs %d", first.Snapshot().Year, second.Snapshot().Year)
	}
	if len(first.Snapshot().TerrainMap.Terrain) != len(second.Snapshot().TerrainMap.Terrain) {
		t.Fatalf("terrain length mismatch: %d vs %d", len(first.Snapshot().TerrainMap.Terrain), len(second.Snapshot().TerrainMap.Terrain))
	}
}

func TestSimulation_Step_AdvancesExactlyOneYear(t *testing.T) {
	sim, err := NewSimulation("gui-step", PopulationParameters{
		GrowthRate:      0.025,
		StarvationRate:  0.10,
		MigrationRate:   0.07,
		MaxCellCapacity: 30,
	})
	if err != nil {
		t.Fatalf("NewSimulation returned error: %v", err)
	}

	before := sim.Snapshot()
	result := sim.Step()
	after := sim.Snapshot()

	if before.Year+1 != after.Year {
		t.Fatalf("expected year to advance by 1, before=%d after=%d", before.Year, after.Year)
	}
	if result.Year != after.Year {
		t.Fatalf("step result year=%d snapshot year=%d", result.Year, after.Year)
	}
	if len(after.TerrainMap.Terrain) != before.TerrainMap.Width*before.TerrainMap.Height {
		t.Fatalf("terrain snapshot length mismatch: got %d want %d", len(after.TerrainMap.Terrain), before.TerrainMap.Width*before.TerrainMap.Height)
	}
}

func TestSimulationSnapshot_ContainsImmutableCellValues(t *testing.T) {
	sim, err := NewSimulation("snapshot-check", PopulationParameters{
		GrowthRate:      0.025,
		StarvationRate:  0.10,
		MigrationRate:   0.07,
		MaxCellCapacity: 30,
	})
	if err != nil {
		t.Fatalf("NewSimulation returned error: %v", err)
	}

	snap := sim.Snapshot()
	if snap.TerrainMap.Width <= 0 || snap.TerrainMap.Height <= 0 {
		t.Fatal("expected terrain snapshot dimensions to be positive")
	}
	if len(snap.TerrainMap.Elevation) != len(snap.TerrainMap.Terrain) {
		t.Fatalf("elevation length mismatch: got %d want %d", len(snap.TerrainMap.Elevation), len(snap.TerrainMap.Terrain))
	}
	if len(snap.Population.Population) != len(snap.TerrainMap.Terrain) {
		t.Fatalf("population length mismatch: got %d want %d", len(snap.Population.Population), len(snap.TerrainMap.Terrain))
	}
}
