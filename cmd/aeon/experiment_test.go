package main

import (
	"testing"

	"github.com/julianstephens/aeon/internal/simulation"
)

func TestBuildExperimentConfig_DefaultConfiguration(t *testing.T) {
	command := ExperimentCommand{
		Seed:     "default_world_001",
		Years:    100,
		Interval: 25,
		Scenario: string(simulation.ScenarioBaseline),
		Format:   "text",
	}

	scenario, config, err := buildExperimentConfig(command)
	if err != nil {
		t.Fatalf("buildExperimentConfig returned error: %v", err)
	}

	if scenario != simulation.ScenarioBaseline {
		t.Fatalf("scenario = %s, want %s", scenario, simulation.ScenarioBaseline)
	}
	if config != simulation.DefaultExperimentConfig() {
		t.Fatalf("config = %+v, want %+v", config, simulation.DefaultExperimentConfig())
	}
}

func TestBuildExperimentConfig_ScenarioConfiguration(t *testing.T) {
	command := ExperimentCommand{
		Seed:     "default_world_001",
		Years:    100,
		Interval: 25,
		Scenario: string(simulation.ScenarioCrowded),
		Format:   "text",
	}

	_, config, err := buildExperimentConfig(command)
	if err != nil {
		t.Fatalf("buildExperimentConfig returned error: %v", err)
	}

	if config.InitialPopulation != 10000 {
		t.Fatalf("initial population = %.0f, want 10000", config.InitialPopulation)
	}
	if config.MigrationRate != 0.07 {
		t.Fatalf("migration rate = %.3f, want 0.070", config.MigrationRate)
	}
}

func TestBuildExperimentConfig_CLIOverridesScenario(t *testing.T) {
	population := 321.0
	migrationRate := 0.42
	maxCellCapacity := 12.0

	command := ExperimentCommand{
		Seed:            "seed-123",
		Years:           30,
		Interval:        10,
		Scenario:        string(simulation.ScenarioScarcity),
		Format:          "json",
		Population:      &population,
		MigrationRate:   &migrationRate,
		MaxCellCapacity: &maxCellCapacity,
	}

	_, config, err := buildExperimentConfig(command)
	if err != nil {
		t.Fatalf("buildExperimentConfig returned error: %v", err)
	}

	if config.Seed != "seed-123" || config.Years != 30 || config.Interval != 10 {
		t.Fatalf("unexpected top-level config overrides: %+v", config)
	}
	if config.InitialPopulation != population {
		t.Fatalf("population override = %.0f, want %.0f", config.InitialPopulation, population)
	}
	if config.MigrationRate != migrationRate {
		t.Fatalf("migration override = %.2f, want %.2f", config.MigrationRate, migrationRate)
	}
	if config.MaxCellCapacity != maxCellCapacity {
		t.Fatalf("capacity override = %.0f, want %.0f", config.MaxCellCapacity, maxCellCapacity)
	}
}

func TestBuildExperimentConfig_InvalidScenarioRejected(t *testing.T) {
	command := ExperimentCommand{Scenario: "unknown"}
	_, _, err := buildExperimentConfig(command)
	if err == nil {
		t.Fatal("expected invalid scenario error")
	}
}

func TestValidateExperimentConfig_InvalidParametersRejected(t *testing.T) {
	config := simulation.DefaultExperimentConfig()
	config.Interval = 101
	if err := simulation.ValidateExperimentConfig(config); err == nil {
		t.Fatal("expected interval > years to be rejected")
	}

	config = simulation.DefaultExperimentConfig()
	config.MigrationRate = 1.1
	if err := simulation.ValidateExperimentConfig(config); err == nil {
		t.Fatal("expected migration-rate > 1 to be rejected")
	}
}
