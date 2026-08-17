package main

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/julianstephens/aeon/internal/simulation"
)

func drawInspector(screen *ebiten.Image, rect image.Rectangle, snapshot simulation.SimulationSnapshot, selected *CellSelection) {
	ebitenutil.DebugPrintAt(screen, "INSPECTOR", rect.Min.X+18, rect.Min.Y+18)
	if selected == nil {
		ebitenutil.DebugPrintAt(screen, "CELL", rect.Min.X+18, rect.Min.Y+50)
		ebitenutil.DebugPrintAt(screen, "Click a cell in the world", rect.Min.X+18, rect.Min.Y+72)
		ebitenutil.DebugPrintAt(screen, "to inspect its state.", rect.Min.X+18, rect.Min.Y+90)
		printWorldStats(screen, rect, snapshot)
		return
	}

	cellX := selected.X
	cellY := selected.Y
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

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("CELL (%d, %d)", cellX, cellY), rect.Min.X+18, rect.Min.Y+50)
	ebitenutil.DebugPrintAt(screen, "TERRAIN", rect.Min.X+18, rect.Min.Y+84)
	ebitenutil.DebugPrintAt(screen, "  "+cellTerrain.String(), rect.Min.X+18, rect.Min.Y+102)

	ebitenutil.DebugPrintAt(screen, "ENVIRONMENT", rect.Min.X+18, rect.Min.Y+138)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Elevation   %.2f", snapshot.TerrainMap.Elevation[index]), rect.Min.X+18, rect.Min.Y+158)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Moisture    %.2f", snapshot.TerrainMap.Moisture[index]), rect.Min.X+18, rect.Min.Y+176)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Fertility   %.2f", snapshot.TerrainMap.Fertility[index]), rect.Min.X+18, rect.Min.Y+194)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Food        %.2f", snapshot.TerrainMap.Food[index]), rect.Min.X+18, rect.Min.Y+212)

	ebitenutil.DebugPrintAt(screen, "POPULATION", rect.Min.X+18, rect.Min.Y+248)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Population  %.1f", cellPop), rect.Min.X+18, rect.Min.Y+268)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Capacity    %.1f", cellCap), rect.Min.X+18, rect.Min.Y+286)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Pressure    %.1f%%", pressure), rect.Min.X+18, rect.Min.Y+304)

	printWorldStats(screen, rect, snapshot)
}

func printWorldStats(screen *ebiten.Image, rect image.Rectangle, snapshot simulation.SimulationSnapshot) {
	y := rect.Max.Y - 160
	ebitenutil.DebugPrintAt(screen, "WORLD", rect.Min.X+18, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Year          %d", snapshot.Year), rect.Min.X+18, y+20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Population    %,.0f", snapshot.Population.TotalPopulation), rect.Min.X+18, y+40)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Capacity      %,.0f", totalCapacity(snapshot.Population.CarryingCapacity)), rect.Min.X+18, y+60)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Utilization   %.1f%%", snapshot.Population.Utilization*100), rect.Min.X+18, y+80)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Occupied      %d", snapshot.Population.OccupiedCells), rect.Min.X+18, y+100)
}
