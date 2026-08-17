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
}
