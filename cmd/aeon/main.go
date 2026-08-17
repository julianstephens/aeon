package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/julianstephens/go-utils/logger"
)

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("aeon"),
		kong.Description("Run Aeon world simulations."),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
	)
	configureLogLevel(cli.LogLevel)

	var err error
	switch ctx.Command() {
	case "run":
		err = runSimulation(cli.Run)
	case "experiment":
		err = runExperiment(cli.Experiment)
	default:
		err = fmt.Errorf("unsupported command: %s", ctx.Command())
	}

	if err != nil {
		logger.Errorf("command failed: %v", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func configureLogLevel(level string) {
	if err := logger.SetLogLevel(level); err != nil {
		logger.Warnf("invalid log level %q, falling back to info", level)
		_ = logger.SetLogLevel("info")
	}
}
