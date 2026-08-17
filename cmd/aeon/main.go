package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/julianstephens/aeon/internal/simulation"
	"github.com/julianstephens/go-utils/cliutil"
	"github.com/julianstephens/go-utils/logger"
)

type CLI struct {
	Seed     string `help:"Seed for world generation." default:"default_world_001"`
	Years    int    `help:"Number of years to simulate." default:"1"`
	LogLevel string `help:"Log level." default:"info" env:"AEON_LOG_LEVEL"`
}

func main() {
	var cli CLI
	kong.Parse(&cli,
		kong.Name("aeon"),
		kong.Description("Run Aeon world simulations."),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
	)
	cliutil.SetDefaultConsole(cliutil.NewConsole(cliutil.NewColoredFormatter(), os.Stdout))
	c := cliutil.DefaultConsole()
	configureLogLevel(cli.LogLevel)

	logger.WithFields(map[string]any{
		"seed":  cli.Seed,
		"years": cli.Years,
	}).Info("starting simulation run")
	_ = c.Info(fmt.Sprintf("Generating world with seed '%s' for %d years...", cli.Seed, cli.Years))

	err := simulation.Run(c, cli.Seed, cli.Years)
	if err != nil {
		logger.Errorf("error running simulation: %v", err)
		_ = c.Error(fmt.Sprintf("Error running simulation: %v", err))
		os.Exit(1)
	}

	logger.Info("simulation completed")
}

func configureLogLevel(level string) {
	if err := logger.SetLogLevel(level); err != nil {
		logger.Warnf("invalid log level %q, falling back to info", level)
		_ = logger.SetLogLevel("info")
	}
}
