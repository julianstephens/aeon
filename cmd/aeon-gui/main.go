package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/julianstephens/aeon/internal/simulation"
)

type CLI struct {
	Seed            string  `help:"Seed for world generation." default:"default_world_001"`
	Population      float64 `help:"Initial population override." default:"1000"`
	GrowthRate      float64 `help:"Growth rate override." default:"0.025"`
	StarvationRate  float64 `help:"Starvation rate override." default:"0.10"`
	MigrationRate   float64 `help:"Migration rate override." default:"0.07"`
	MaxCellCapacity float64 `help:"Maximum cell capacity override." default:"30"`
	Years           int     `help:"Years to advance on launch." default:"0"`
}

func main() {
	var cli CLI
	kong.Parse(&cli,
		kong.Name("aeon-gui"),
		kong.Description("Aeon simulation inspector."),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
	)

	cfg := simulation.DefaultGUIConfig()
	cfg.Seed = cli.Seed
	cfg.InitialPopulation = cli.Population
	cfg.GrowthRate = cli.GrowthRate
	cfg.StarvationRate = cli.StarvationRate
	cfg.MigrationRate = cli.MigrationRate
	cfg.MaxCellCapacity = cli.MaxCellCapacity

	app, err := NewApp(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for range cli.Years {
		app.simulation.Step()
		app.snapshot = app.simulation.Snapshot()
	}

	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Aeon")
	if err := ebiten.RunGame(app); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
