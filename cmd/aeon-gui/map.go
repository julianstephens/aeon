package main

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation"
)

type LayerMode int

const (
	LayerTerrain LayerMode = iota
	LayerElevation
	LayerMoisture
	LayerFertility
	LayerFood
	LayerPopulation
)

var layerNames = []string{"Terrain", "Elevation", "Moisture", "Fertility", "Food", "Population"}

type CellSelection struct {
	X int
	Y int
}

type layerShortcut struct {
	key  ebiten.Key
	mode LayerMode
}

var layerShortcuts = []layerShortcut{
	{key: ebiten.Key1, mode: LayerTerrain},
	{key: ebiten.Key2, mode: LayerElevation},
	{key: ebiten.Key3, mode: LayerMoisture},
	{key: ebiten.Key4, mode: LayerFertility},
	{key: ebiten.Key5, mode: LayerFood},
	{key: ebiten.Key6, mode: LayerPopulation},
}

func drawMap(screen *ebiten.Image, snapshot simulation.SimulationSnapshot, mode LayerMode, viewport image.Rectangle, selected *CellSelection) {
	if snapshot.TerrainMap.Width == 0 || snapshot.TerrainMap.Height == 0 {
		return
	}
	cellW := float64(viewport.Dx()) / float64(snapshot.TerrainMap.Width)
	cellH := float64(viewport.Dy()) / float64(snapshot.TerrainMap.Height)
	maxPop := maxFloat64(snapshot.Population.Population)
	for index, terrain := range snapshot.TerrainMap.Terrain {
		x := index % snapshot.TerrainMap.Width
		y := index / snapshot.TerrainMap.Width
		startX := viewport.Min.X + int(float64(x)*cellW)
		startY := viewport.Min.Y + int(float64(y)*cellH)
		endX := viewport.Min.X + int(float64(x+1)*cellW)
		endY := viewport.Min.Y + int(float64(y+1)*cellH)
		img := ebiten.NewImage(endX-startX, endY-startY)
		img.Fill(colorForCell(snapshot, index, mode, maxPop, terrain))
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(startX), float64(startY))
		screen.DrawImage(img, opts)
	}
	if selected != nil {
		cellX := viewport.Min.X + int(float64(selected.X)*cellW)
		cellY := viewport.Min.Y + int(float64(selected.Y)*cellH)
		outline := ebiten.NewImage(int(cellW)+2, int(cellH)+2)
		outline.Fill(color.RGBA{255, 255, 255, 255})
		if cellX >= viewport.Min.X && cellX+int(cellW) <= viewport.Max.X && cellY >= viewport.Min.Y && cellY+int(cellH) <= viewport.Max.Y {
			outlineOpts := &ebiten.DrawImageOptions{}
			outlineOpts.GeoM.Translate(float64(cellX-1), float64(cellY-1))
			screen.DrawImage(outline, outlineOpts)
		}
	}
	label := "World"
	ebitenutil.DebugPrintAt(screen, label, viewport.Min.X+10, viewport.Min.Y+10)
}

func colorForCell(snapshot simulation.SimulationSnapshot, index int, mode LayerMode, maxPop float64, terrain simtypes.TerrainType) color.RGBA {
	if mode == LayerTerrain {
		switch terrain {
		case simtypes.TerrainTypeWater:
			return color.RGBA{R: 38, G: 104, B: 181, A: 255}
		case simtypes.TerrainTypePlains:
			return color.RGBA{R: 104, G: 160, B: 82, A: 255}
		case simtypes.TerrainTypeForest:
			return color.RGBA{R: 46, G: 110, B: 58, A: 255}
		case simtypes.TerrainTypeMountain:
			return color.RGBA{R: 145, G: 145, B: 145, A: 255}
		default:
			return color.RGBA{R: 40, G: 40, B: 40, A: 255}
		}
	}
	var value float64
	switch mode {
	case LayerElevation:
		value = snapshot.TerrainMap.Elevation[index]
	case LayerMoisture:
		value = snapshot.TerrainMap.Moisture[index]
	case LayerFertility:
		value = snapshot.TerrainMap.Fertility[index]
	case LayerFood:
		value = snapshot.TerrainMap.Food[index]
	case LayerPopulation:
		value = snapshot.Population.Population[index]
		if maxPop <= 0 {
			return color.RGBA{R: 20, G: 20, B: 20, A: 255}
		}
		value = value / maxPop
		if value <= 0 {
			return color.RGBA{R: 12, G: 12, B: 12, A: 255}
		}
		if value > 1 {
			value = 1
		}
		v := uint8(20 + 235*value)
		return color.RGBA{R: v, G: v, B: v, A: 255}
	default:
		value = 0
	}
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	v := uint8(20 + 235*value)
	return color.RGBA{R: v, G: v, B: v, A: 255}
}

func maxFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func normalizeMinMax(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	minV := values[0]
	maxV := values[0]
	for _, v := range values {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	if math.Abs(maxV-minV) < 1e-9 {
		out := make([]float64, len(values))
		for i := range out {
			out[i] = 0.5
		}
		return out
	}
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = clamp01((v - minV) / (maxV - minV))
	}
	return out
}
