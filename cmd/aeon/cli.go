package main

import (
	"fmt"
	"os"

	"github.com/julianstephens/aeon/internal/simulation"
	"github.com/julianstephens/go-utils/cliutil"
	"github.com/julianstephens/go-utils/logger"
)

type CLI struct {
	LogLevel   string            `help:"Log level."                   default:"info"     env:"AEON_LOG_LEVEL"`
	Run        RunCommand        `help:"Run a standard simulation."   default:"withargs"                      cmd:""`
	Experiment ExperimentCommand `help:"Run a population experiment."                                         cmd:""`
}

type RunCommand struct {
	Seed  string `help:"Seed for world generation."   default:"default_world_001"`
	Years int    `help:"Number of years to simulate." default:"1"`
}

type ExperimentCommand struct {
	Seed            string   `help:"Seed for world generation."                             default:"default_world_001"`
	Years           int      `help:"Number of years to simulate."                           default:"100"`
	Population      *float64 `help:"Initial population override for the selected scenario."                             name:"population"`
	GrowthRate      *float64 `help:"Growth rate override for the selected scenario."                                    name:"growth-rate"`
	StarvationRate  *float64 `help:"Starvation rate override for the selected scenario."                                name:"starvation-rate"`
	MigrationRate   *float64 `help:"Migration rate override for the selected scenario."                                 name:"migration-rate"`
	MaxCellCapacity *float64 `help:"Max cell capacity override for the selected scenario."                              name:"max-cell-capacity"`
	Interval        int      `help:"Sampling interval in years."                            default:"25"`
	Scenario        string   `help:"Scenario preset name."                                  default:"baseline"`
	Format          string   `help:"Output format."                                         default:"text"                                       enum:"text,json"`
}

func runSimulation(command RunCommand) error {
	cliutil.SetDefaultConsole(cliutil.NewConsole(cliutil.NewColoredFormatter(), os.Stdout))
	console := cliutil.DefaultConsole()

	logger.WithFields(map[string]any{
		"seed":  command.Seed,
		"years": command.Years,
	}).Info("starting simulation run")
	_ = console.Info(fmt.Sprintf("Generating world with seed '%s' for %d years...", command.Seed, command.Years))

	if err := simulation.Run(console, command.Seed, command.Years); err != nil {
		logger.Errorf("error running simulation: %v", err)
		_ = console.Error(fmt.Sprintf("Error running simulation: %v", err))
		return err
	}
	return nil
}
