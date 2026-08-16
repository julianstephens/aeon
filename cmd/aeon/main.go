package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/julianstephens/aeon/internal/simulation"
	"github.com/julianstephens/go-utils/cliutil"
)

func main() {
	args := cliutil.ParseArgs(os.Args)
	cliutil.SetDefaultConsole(cliutil.NewConsole(cliutil.NewColoredFormatter(), os.Stdout))
	c := cliutil.DefaultConsole()

	seed := args.GetFlag("--seed")
	if seed == "" {
		seed = "default_world_001"
	}

	years := args.GetFlagWithDefault("--years", "1")
	yearsNum, err := strconv.Atoi(years)
	if err != nil {
		_ = c.Error(fmt.Sprintf("Invalid years flag: %s", years))
		os.Exit(1)
	}

	_ = c.Info(fmt.Sprintf("Generating world with seed '%s' for %d years...", seed, yearsNum))

	err = simulation.Run(c, seed, yearsNum)
	if err != nil {
		_ = c.Error(fmt.Sprintf("Error running simulation: %v", err))
		os.Exit(1)
	}
}
