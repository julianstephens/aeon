package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/julianstephens/aeon/internal/simulation"
)

func TestRenderExperimentText_ContainsExpectedMetrics(t *testing.T) {
	result := simulation.ExperimentResult{
		Config: simulation.ExperimentConfig{
			Seed:              "42",
			Years:             100,
			InitialPopulation: 1000,
			Interval:          25,
		},
		Samples: []simulation.PopulationSample{
			{Year: 0, Population: 1000, CarryingCapacity: 8000, Utilization: 0.125, OccupiedCells: 20},
			{
				Year:               25,
				Population:         1100,
				CarryingCapacity:   8000,
				Utilization:        0.1375,
				OccupiedCells:      24,
				MigratedPopulation: 41,
				StarvationDeaths:   3,
			},
		},
	}

	output := renderExperimentText(simulation.ScenarioBaseline, result)

	if !strings.Contains(output, "Aeon Population Experiment") {
		t.Fatalf("expected title in text output, got %q", output)
	}
	if !strings.Contains(output, "Year") || !strings.Contains(output, "Population") ||
		!strings.Contains(output, "Migrants") {
		t.Fatalf("expected table headers in text output, got %q", output)
	}
	if !strings.Contains(output, "Final") {
		t.Fatalf("expected final summary in text output, got %q", output)
	}
}

func TestRenderExperimentJSON_SerializesResult(t *testing.T) {
	result := simulation.ExperimentResult{
		Config:  simulation.DefaultExperimentConfig(),
		Samples: []simulation.PopulationSample{{Year: 0, Population: 1000}},
	}

	payload, err := renderExperimentJSON(result)
	if err != nil {
		t.Fatalf("renderExperimentJSON returned error: %v", err)
	}

	var decoded simulation.ExperimentResult
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("failed to decode JSON payload: %v", err)
	}

	if len(decoded.Samples) != 1 || decoded.Samples[0].Year != 0 {
		t.Fatalf("unexpected decoded samples: %+v", decoded.Samples)
	}
}

func TestRenderExperimentJSON_IsDeterministic(t *testing.T) {
	result := simulation.ExperimentResult{
		Config: simulation.DefaultExperimentConfig(),
		Samples: []simulation.PopulationSample{
			{Year: 0, Population: 1000},
			{Year: 25, Population: 1050, MigratedPopulation: 12, StarvationDeaths: 1},
		},
	}

	first, err := renderExperimentJSON(result)
	if err != nil {
		t.Fatalf("first renderExperimentJSON returned error: %v", err)
	}

	second, err := renderExperimentJSON(result)
	if err != nil {
		t.Fatalf("second renderExperimentJSON returned error: %v", err)
	}

	if first != second {
		t.Fatalf("expected deterministic JSON output, got\n%s\n---\n%s", first, second)
	}
}
