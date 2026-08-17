package main

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/julianstephens/aeon/internal/simulation"
)

func drawInspector(screen *ebiten.Image, rect image.Rectangle, snapshot simulation.SimulationSnapshot, selected *CellSelection) {
	if selected == nil {
		ebitenutil.DebugPrintAt(screen, "Click a cell to inspect it.", rect.Min.X+12, rect.Min.Y+16)
		ebitenutil.DebugPrintAt(screen, "", rect.Min.X+12, rect.Min.Y+40)
		printWorldStats(screen, rect, snapshot)
		return
	}
	cellX := selected.X
	cellY := selected.Y
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Cell (%d, %d)", cellX, cellY), rect.Min.X+12, rect.Min.Y+16)
	if cellX < 0 || cellX >= snapshot.TerrainMap.Width || cellY < 0 || cellY >= snapshot.TerrainMap.Height {
		return
	}
	index := cellY*snapshot.TerrainMap.Width + cellX
	cellTerrain := snapshot.TerrainMap.Terrain[index]
	cellPop := snapshot.Population.Population[index]
	cellCap := snapshot.Population.CarryingCapacity[index]
	pressure := 0.0
	if cellCap > 0 {
		pressure = (cellPop / cellCap) * 100
	}
	ebitenutil.DebugPrintAt(screen, "Terrain", rect.Min.X+12, rect.Min.Y+48)
	ebitenutil.DebugPrintAt(screen, "  "+cellTerrain.String(), rect.Min.X+12, rect.Min.Y+68)
	ebitenutil.DebugPrintAt(screen, "Environment", rect.Min.X+12, rect.Min.Y+96)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Elevation  %.2f", snapshot.TerrainMap.Elevation[index]), rect.Min.X+12, rect.Min.Y+116)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Moisture   %.2f", snapshot.TerrainMap.Moisture[index]), rect.Min.X+12, rect.Min.Y+136)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Fertility  %.2f", snapshot.TerrainMap.Fertility[index]), rect.Min.X+12, rect.Min.Y+156)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Food       %.2f", snapshot.TerrainMap.Food[index]), rect.Min.X+12, rect.Min.Y+176)
	ebitenutil.DebugPrintAt(screen, "Population", rect.Min.X+12, rect.Min.Y+208)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Population %.1f", cellPop), rect.Min.X+12, rect.Min.Y+228)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Capacity   %.1f", cellCap), rect.Min.X+12, rect.Min.Y+248)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Pressure   %.1f%%", pressure), rect.Min.X+12, rect.Min.Y+268)
	printWorldStats(screen, rect, snapshot)
}

func printWorldStats(screen *ebiten.Image, rect image.Rectangle, snapshot simulation.SimulationSnapshot) {
	y := rect.Min.Y + 320
	ebitenutil.DebugPrintAt(screen, "Summary", rect.Min.X+12, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Year:         %d", snapshot.Year), rect.Min.X+12, y+20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Population:   %.0f", snapshot.Population.TotalPopulation), rect.Min.X+12, y+40)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Capacity:     %.0f", totalCapacity(snapshot.Population.CarryingCapacity)), rect.Min.X+12, y+60)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Utilization:  %.1f%%", snapshot.Population.Utilization*100), rect.Min.X+12, y+80)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Occupied:     %d", snapshot.Population.OccupiedCells), rect.Min.X+12, y+100)
}
