package main

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/julianstephens/aeon/internal/simulation"
)

func drawHeader(screen *ebiten.Image, snapshot simulation.SimulationSnapshot, seed string) {
	label := fmt.Sprintf("Aeon                         Year %d       Seed: %s", snapshot.Year, seed)
	ebitenutil.DebugPrintAt(screen, label, 16, 20)
}

func drawControls(screen *ebiten.Image, rect image.Rectangle, app *App) {
	speedText := fmt.Sprintf("Speed: %.2fx", app.displaySpeed())
	ebitenutil.DebugPrintAt(screen, "◀  Step  ▶  Play/Pause  Reset", rect.Min.X+20, rect.Min.Y+20)
	ebitenutil.DebugPrintAt(screen, speedText, rect.Min.X+420, rect.Min.Y+20)
	if app.playing {
		ebitenutil.DebugPrintAt(screen, "Playing", rect.Min.X+560, rect.Min.Y+20)
	} else {
		ebitenutil.DebugPrintAt(screen, "Paused", rect.Min.X+560, rect.Min.Y+20)
	}
	populationLabel := fmt.Sprintf("Population: %d", int(app.snapshot.Population.TotalPopulation))
	ebitenutil.DebugPrintAt(screen, populationLabel, rect.Min.X+760, rect.Min.Y+20)
	capacityLabel := "Capacity: 0"
	if len(app.snapshot.Population.CarryingCapacity) > 0 {
		capacityLabel = fmt.Sprintf("Capacity: %.0f", totalCapacity(app.snapshot.Population.CarryingCapacity))
	}
	ebitenutil.DebugPrintAt(screen, capacityLabel, rect.Min.X+960, rect.Min.Y+20)
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

func totalCapacity(values []float64) float64 {
	var total float64
	for _, v := range values {
		total += v
	}
	return total
}
