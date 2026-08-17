package simulation

import (
	"reflect"
	"testing"
)

func TestExperimentRun_RecordsYearZero(t *testing.T) {
	config := DefaultExperimentConfig()
	config.Years = 10
	config.Interval = 5
	config.Seed = "experiment-year-zero"

	experiment := NewExperiment(config)
	result, err := experiment.RunE()
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	if len(result.Samples) == 0 {
		t.Fatal("expected at least one sample")
	}

	yearZero := result.Samples[0]
	if yearZero.Year != 0 {
		t.Fatalf("first sample year = %d, want 0", yearZero.Year)
	}
	if yearZero.MigratedPopulation != 0 || yearZero.StarvationDeaths != 0 {
		t.Fatalf(
			"expected year zero interval events to be zero, got migrated=%.2f deaths=%.2f",
			yearZero.MigratedPopulation,
			yearZero.StarvationDeaths,
		)
	}
}

func TestExperimentRun_SamplesAtConfiguredInterval(t *testing.T) {
	config := DefaultExperimentConfig()
	config.Seed = "experiment-interval"
	config.Years = 100
	config.Interval = 25

	experiment := NewExperiment(config)
	result, err := experiment.RunE()
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	wantYears := []int{0, 25, 50, 75, 100}
	gotYears := sampleYears(result)
	if !reflect.DeepEqual(gotYears, wantYears) {
		t.Fatalf("sample years = %v, want %v", gotYears, wantYears)
	}
}

func TestExperimentRun_AlwaysRecordsFinalYear(t *testing.T) {
	config := DefaultExperimentConfig()
	config.Seed = "experiment-final-year"
	config.Years = 100
	config.Interval = 30

	experiment := NewExperiment(config)
	result, err := experiment.RunE()
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	wantYears := []int{0, 30, 60, 90, 100}
	gotYears := sampleYears(result)
	if !reflect.DeepEqual(gotYears, wantYears) {
		t.Fatalf("sample years = %v, want %v", gotYears, wantYears)
	}
}

func TestExperimentRun_PreservesInitialPopulationInYearZeroSample(t *testing.T) {
	config := DefaultExperimentConfig()
	config.Seed = "experiment-initial-pop"
	config.InitialPopulation = 1000
	config.Years = 1
	config.Interval = 1

	experiment := NewExperiment(config)
	result, err := experiment.RunE()
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	if result.Samples[0].Population != config.InitialPopulation {
		t.Fatalf("year zero population = %.0f, want %.0f", result.Samples[0].Population, config.InitialPopulation)
	}
}

func TestExperimentRun_IsDeterministicForSameConfiguration(t *testing.T) {
	config := DefaultExperimentConfig()
	config.Seed = "experiment-deterministic"
	config.Years = 50
	config.Interval = 25

	first, err := NewExperiment(config).RunE()
	if err != nil {
		t.Fatalf("first RunE returned error: %v", err)
	}
	second, err := NewExperiment(config).RunE()
	if err != nil {
		t.Fatalf("second RunE returned error: %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic results, got\n%+v\nvs\n%+v", first, second)
	}
}

func TestExperimentRun_DifferentSeedsProduceDifferentWorlds(t *testing.T) {
	configA := DefaultExperimentConfig()
	configA.Seed = "experiment-seed-a"
	configA.Years = 50
	configA.Interval = 25

	configB := configA
	configB.Seed = "experiment-seed-b"

	first, err := NewExperiment(configA).RunE()
	if err != nil {
		t.Fatalf("first RunE returned error: %v", err)
	}
	second, err := NewExperiment(configB).RunE()
	if err != nil {
		t.Fatalf("second RunE returned error: %v", err)
	}

	if reflect.DeepEqual(first.Samples, second.Samples) {
		t.Fatalf("expected different samples for different seeds, got identical results")
	}
}

func sampleYears(result ExperimentResult) []int {
	years := make([]int, 0, len(result.Samples))
	for _, sample := range result.Samples {
		years = append(years, sample.Year)
	}
	return years
}
