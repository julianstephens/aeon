package main

import (
	"fmt"
	"strings"

	"github.com/julianstephens/aeon/internal/simulation"
)

func runExperiment(command ExperimentCommand) error {
	scenario, config, err := buildExperimentConfig(command)
	if err != nil {
		return err
	}

	if err := simulation.ValidateExperimentConfig(config); err != nil {
		return err
	}

	experiment := simulation.NewExperiment(config)
	result, err := experiment.RunE()
	if err != nil {
		return err
	}

	switch strings.ToLower(command.Format) {
	case "text":
		fmt.Print(renderExperimentText(scenario, result))
	case "json":
		payload, jsonErr := renderExperimentJSON(result)
		if jsonErr != nil {
			return jsonErr
		}
		fmt.Println(payload)
	default:
		return fmt.Errorf("unsupported format: %s", command.Format)
	}

	return nil
}

func buildExperimentConfig(command ExperimentCommand) (simulation.Scenario, simulation.ExperimentConfig, error) {
	scenario, err := simulation.ParseScenario(command.Scenario)
	if err != nil {
		return "", simulation.ExperimentConfig{}, err
	}

	config := simulation.ScenarioConfig(scenario)
	config.Seed = command.Seed
	config.Years = command.Years
	config.Interval = command.Interval

	if command.Population != nil {
		config.InitialPopulation = *command.Population
	}
	if command.GrowthRate != nil {
		config.GrowthRate = *command.GrowthRate
	}
	if command.StarvationRate != nil {
		config.StarvationRate = *command.StarvationRate
	}
	if command.MigrationRate != nil {
		config.MigrationRate = *command.MigrationRate
	}
	if command.MaxCellCapacity != nil {
		config.MaxCellCapacity = *command.MaxCellCapacity
	}

	return scenario, config, nil
}
