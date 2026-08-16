package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/julianstephens/aeon/internal/simulation"
	"github.com/julianstephens/go-utils/cliutil"
	"github.com/julianstephens/go-utils/logger"
)

func main() {
	args := cliutil.ParseArgs(os.Args)
	cliutil.SetDefaultConsole(cliutil.NewConsole(cliutil.NewColoredFormatter(), os.Stdout))
	c := cliutil.DefaultConsole()
	configureLogLevel(args)

	seed := args.GetFlag("--seed")
	if seed == "" {
		seed = "default_world_001"
	}

	years := args.GetFlagWithDefault("--years", "1")
	yearsNum, err := strconv.Atoi(years)
	if err != nil {
		logger.Errorf("invalid years flag value: %s", years)
		_ = c.Error(fmt.Sprintf("Invalid years flag: %s", years))
		os.Exit(1)
	}

	logger.WithFields(map[string]interface{}{
		"seed":  seed,
		"years": yearsNum,
	}).Info("starting simulation run")
	_ = c.Info(fmt.Sprintf("Generating world with seed '%s' for %d years...", seed, yearsNum))

	err = simulation.Run(c, seed, yearsNum)
	if err != nil {
		logger.Errorf("error running simulation: %v", err)
		_ = c.Error(fmt.Sprintf("Error running simulation: %v", err))
		os.Exit(1)
	}

	logger.Info("simulation completed")
}

func configureLogLevel(args *cliutil.Args) {
	level := args.GetFlag("--log-level")
	if level == "" {
		level = os.Getenv("AEON_LOG_LEVEL")
	}
	if level == "" {
		level = "info"
	}

	if err := logger.SetLogLevel(level); err != nil {
		logger.Warnf("invalid log level %q, falling back to info", level)
		_ = logger.SetLogLevel("info")
	}
}
