package main

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/julianstephens/aeon/internal/simulation"
)

func drawHeader(screen *ebiten.Image, snapshot simulation.SimulationSnapshot) {
	label := fmt.Sprintf("Aeon     Year %d     Seed: %s", snapshot.Year, "42")
	ebitenutil.DebugPrintAt(screen, label, 14, 18)
}

func drawControls(screen *ebiten.Image, rect image.Rectangle, app *App) {
	ebitenutil.DebugPrintAt(screen, "◀ Step  Play  Reset  Speed: 1x", rect.Min.X+20, rect.Min.Y+20)
	if app.playing {
		ebitenutil.DebugPrintAt(screen, "Playing", rect.Min.X+520, rect.Min.Y+20)
	} else {
		ebitenutil.DebugPrintAt(screen, "Paused", rect.Min.X+520, rect.Min.Y+20)
	}
	fmtString := fmt.Sprintf("Population: %d", int(app.snapshot.Population.TotalPopulation))
	ebitenutil.DebugPrintAt(screen, fmtString, rect.Min.X+760, rect.Min.Y+20)
}

func drawLayerList(screen *ebiten.Image, rect image.Rectangle, current LayerMode) {
	for i, item := range layerNames {
		y := rect.Min.Y + 24 + i*32
		txt := item
		if i == int(current) {
			txt = "• " + txt
		}
		ebitenutil.DebugPrintAt(screen, txt, rect.Min.X+18, y)
	}
}
