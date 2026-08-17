package main

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/julianstephens/aeon/internal/simulation"
)

func drawHeader(screen *ebiten.Image, snapshot simulation.SimulationSnapshot, seed string) {
	ebitenutil.DebugPrintAt(screen, "AEON", 18, 18)
	ebitenutil.DebugPrintAt(screen, "SIMULATION INSPECTOR", 18, 34)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("YEAR %d", snapshot.Year), 1040, 18)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SEED %s", seed), 1040, 34)
}

func drawControls(screen *ebiten.Image, rect image.Rectangle, app *App) {
	y := rect.Min.Y + 16
	ebitenutil.DebugPrintAt(screen, "SPACE", rect.Min.X+20, y)
	ebitenutil.DebugPrintAt(screen, "Play / Pause", rect.Min.X+72, y)
	ebitenutil.DebugPrintAt(screen, "RIGHT", rect.Min.X+190, y)
	ebitenutil.DebugPrintAt(screen, "Step", rect.Min.X+240, y)
	ebitenutil.DebugPrintAt(screen, "R", rect.Min.X+300, y)
	ebitenutil.DebugPrintAt(screen, "Reset", rect.Min.X+320, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Speed %.2fx", app.displaySpeed()), rect.Min.X+430, y)

	status := "PAUSED"
	if app.playing {
		status = "PLAYING"
	}
	ebitenutil.DebugPrintAt(screen, status, rect.Min.X+560, y)

	populationLabel := fmt.Sprintf("Population %,.0f", app.snapshot.Population.TotalPopulation)
	capacityLabel := fmt.Sprintf("Capacity %,.0f", totalCapacity(app.snapshot.Population.CarryingCapacity))
	utilizationLabel := fmt.Sprintf("Utilization %.1f%%", app.snapshot.Population.Utilization*100)
	ebitenutil.DebugPrintAt(screen, populationLabel, rect.Min.X+760, y)
	ebitenutil.DebugPrintAt(screen, capacityLabel, rect.Min.X+930, y)
	ebitenutil.DebugPrintAt(screen, utilizationLabel, rect.Min.X+1080, y)

	ebitenutil.DebugPrintAt(screen, "S cycles speed", rect.Min.X+430, rect.Min.Y+42)
	ebitenutil.DebugPrintAt(screen, "1–6 layers", rect.Min.X+560, rect.Min.Y+42)
}

func drawLayerList(screen *ebiten.Image, rect image.Rectangle, current LayerMode) {
	ebitenutil.DebugPrintAt(screen, "LAYERS", rect.Min.X+18, rect.Min.Y+18)
	ebitenutil.DebugPrintAt(screen, "────────────────", rect.Min.X+18, rect.Min.Y+34)
	for i, item := range layerNames {
		y := rect.Min.Y + 62 + i*42
		marker := "  "
		if i == int(current) {
			marker = "> "
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%d  %s", marker, i+1, item), rect.Min.X+18, y)
	}
	ebitenutil.DebugPrintAt(screen, "MAP", rect.Min.X+18, rect.Max.Y-82)
	ebitenutil.DebugPrintAt(screen, "Click a cell to inspect", rect.Min.X+18, rect.Max.Y-62)
	ebitenutil.DebugPrintAt(screen, "current world state", rect.Min.X+18, rect.Max.Y-42)
}

func totalCapacity(values []float64) float64 {
	var total float64
	for _, v := range values {
		total += v
	}
	return total
}
