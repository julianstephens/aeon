package simulation

import "fmt"

type Scenario string

const (
	ScenarioBaseline Scenario = "baseline"
	ScenarioCrowded  Scenario = "crowded"
	ScenarioIsolated Scenario = "isolated"
	ScenarioScarcity Scenario = "scarcity"
)

func ParseScenario(name string) (Scenario, error) {
	scenario := Scenario(name)
	switch scenario {
	case ScenarioBaseline, ScenarioCrowded, ScenarioIsolated, ScenarioScarcity:
		return scenario, nil
	default:
		return "", fmt.Errorf(
			"invalid scenario %q (supported: %s, %s, %s, %s)",
			name,
			ScenarioBaseline,
			ScenarioCrowded,
			ScenarioIsolated,
			ScenarioScarcity,
		)
	}
}

func ScenarioConfig(name Scenario) ExperimentConfig {
	config := DefaultExperimentConfig()

	switch name {
	case ScenarioCrowded:
		config.InitialPopulation = 10000
	case ScenarioIsolated:
		config.InitialPopulation = 10000
		config.MigrationRate = 0
	case ScenarioScarcity:
		config.InitialPopulation = 10000
		config.MaxCellCapacity = 10
	case ScenarioBaseline:
		fallthrough
	default:
		// Baseline already represented by defaults.
	}

	return config
}
