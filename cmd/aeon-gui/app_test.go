package main

import (
	"testing"

	"github.com/julianstephens/aeon/internal/simulation"
)

func TestAppResetRestoresInitialState(t *testing.T) {
	cfg := simulation.DefaultGUIConfig()
	cfg.Seed = "app-reset-test"
	cfg.InitialPopulation = 1200
	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	app.playing = true
	app.selectedCell = &CellSelection{X: 5, Y: 7}
	app.simulation.Step()
	if err := app.reset(); err != nil {
		t.Fatalf("reset returned error: %v", err)
	}
	if app.playing {
		t.Fatal("expected reset to stop playback")
	}
	if app.selectedCell != nil {
		t.Fatal("expected reset to clear selected cell")
	}
	if app.simulation == nil || app.simulation.World() == nil {
		t.Fatal("reset should recreate simulation state")
	}
	if app.snapshot.Year != 0 {
		t.Fatalf("after reset year = %d, want 0", app.snapshot.Year)
	}
}

func TestAppSpeedSettingsCycle(t *testing.T) {
	app, err := NewApp(simulation.DefaultGUIConfig())
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	initial := app.yearsPerSecond
	app.cycleSpeed(1)
	if app.yearsPerSecond == initial {
		t.Fatal("expected speed to change after cycle")
	}
	if app.yearsPerSecond != app.displaySpeed() {
		t.Fatal("speed display should match runtime speed")
	}
}

func TestAppResetReturnsToDefaultSpeed(t *testing.T) {
	app, err := NewApp(simulation.DefaultGUIConfig())
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	app.yearsPerSecond = 10
	app.playing = true
	if err := app.reset(); err != nil {
		t.Fatalf("reset returned error: %v", err)
	}
	if app.yearsPerSecond != 1 {
		t.Fatalf("reset yearsPerSecond = %v, want 1", app.yearsPerSecond)
	}
	if app.playing {
		t.Fatal("expected reset to stop playback")
	}
}
