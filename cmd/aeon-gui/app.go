package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/julianstephens/aeon/internal/simulation"
)

const (
	windowWidth  = 1280
	windowHeight = 800
)

type App struct {
	simulation      *simulation.Simulation
	config          simulation.GUIConfig
	snapshot        simulation.SimulationSnapshot
	layer           LayerMode
	playing         bool
	yearsPerSecond  float64
	selectedCell    *CellSelection
	accumulatedStep float64
}

func NewApp(cfg simulation.GUIConfig) (*App, error) {
	sim, err := simulation.NewSimulationWithConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &App{
		simulation:     sim,
		config:         cfg,
		snapshot:       sim.Snapshot(),
		layer:          LayerTerrain,
		yearsPerSecond: 1,
		selectedCell:   nil,
	}, nil
}

func (a *App) Update() error {
	if a == nil || a.simulation == nil {
		return nil
	}
	if a.yearsPerSecond <= 0 {
		a.yearsPerSecond = 1
	}

	if a.playing {
		a.accumulatedStep += 1.0 / float64(ebiten.TPS())
		for a.accumulatedStep >= 1.0/a.yearsPerSecond {
			a.simulation.Step()
			a.accumulatedStep -= 1.0 / a.yearsPerSecond
		}
		a.snapshot = a.simulation.Snapshot()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		a.playing = !a.playing
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		a.simulation.Step()
		a.snapshot = a.simulation.Snapshot()
		a.playing = false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if err := a.reset(); err != nil {
			return err
		}
	}
	for i := range layerShortcuts {
		if inpututil.IsKeyJustPressed(layerShortcuts[i].key) {
			a.layer = layerShortcuts[i].mode
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if cell, ok := a.mapCellFromMouse(); ok {
			a.selectedCell = cell
		}
	}
	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	screen.Fill(themeBackground)

	headerRect := image.Rect(0, 0, windowWidth, 64)
	drawPanel(screen, headerRect, themePanel, 2)
	drawHeader(screen, a.snapshot)

	sidebarRect := image.Rect(0, 64, 240, 720)
	drawPanel(screen, sidebarRect, themePanel, 2)
	drawLayerList(screen, sidebarRect, a.layer)

	mapRect := image.Rect(240, 64, 1040, 720)
	drawPanel(screen, mapRect, themePanel, 2)
	drawMap(screen, a.snapshot, a.layer, mapRect, a.selectedCell)

	inspectorRect := image.Rect(1040, 64, 1280, 720)
	drawPanel(screen, inspectorRect, themePanel, 2)
	drawInspector(screen, inspectorRect, a.snapshot, a.selectedCell)

	controlsRect := image.Rect(0, 720, windowWidth, windowHeight)
	drawPanel(screen, controlsRect, themePanel, 2)
	drawControls(screen, controlsRect, a)
}

func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowWidth, windowHeight
}

func (a *App) reset() error {
	cfg := a.config
	rebuilt, err := simulation.NewSimulationWithConfig(cfg)
	if err != nil {
		return err
	}
	a.simulation = rebuilt
	a.snapshot = rebuilt.Snapshot()
	a.playing = false
	a.accumulatedStep = 0
	a.selectedCell = nil
	return nil
}

func (a *App) mapCellFromMouse() (*CellSelection, bool) {
	posX, posY := ebiten.CursorPosition()
	viewport := image.Rect(240, 64, 1040, 720)
	if !pointInRect(posX, posY, viewport) {
		return nil, false
	}
	mapW, mapH := a.snapshot.TerrainMap.Width, a.snapshot.TerrainMap.Height
	innerX := posX - viewport.Min.X
	innerY := posY - viewport.Min.Y
	cellW := float64(viewport.Dx()) / float64(mapW)
	cellH := float64(viewport.Dy()) / float64(mapH)
	cellX := int(float64(innerX) / cellW)
	cellY := int(float64(innerY) / cellH)
	if cellX < 0 || cellX >= mapW || cellY < 0 || cellY >= mapH {
		return nil, false
	}
	return &CellSelection{X: cellX, Y: cellY}, true
}

func pointInRect(x, y int, r image.Rectangle) bool {
	return x >= r.Min.X && x < r.Max.X && y >= r.Min.Y && y < r.Max.Y
}

func drawPanel(screen *ebiten.Image, rect image.Rectangle, col color.Color, border int) {
	img := ebiten.NewImage(rect.Dx(), rect.Dy())
	img.Fill(col)
	screen.DrawImage(img, &ebiten.DrawImageOptions{})
	_ = border
}
