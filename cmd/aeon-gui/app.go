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

	headerHeight   = 56
	footerHeight   = 80
	sidebarWidth   = 200
	inspectorWidth = 320
	contentHeight  = windowHeight - headerHeight - footerHeight
	mapPanelWidth  = windowWidth - sidebarWidth - inspectorWidth
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

var speedLevels = []float64{0.25, 0.5, 1, 2, 5, 10}

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
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if err := a.reset(); err != nil {
			return err
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		a.cycleSpeed(1)
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

	headerRect := image.Rect(0, 0, windowWidth, headerHeight)
	drawPanel(screen, headerRect, themePanel, 1)
	drawHeader(screen, a.snapshot, a.config.Seed)

	contentTop := headerHeight
	contentBottom := windowHeight - footerHeight
	sidebarRect := image.Rect(0, contentTop, sidebarWidth, contentBottom)
	mapPanelRect := image.Rect(sidebarWidth, contentTop, windowWidth-inspectorWidth, contentBottom)
	inspectorRect := image.Rect(windowWidth-inspectorWidth, contentTop, windowWidth, contentBottom)

	drawPanel(screen, sidebarRect, themePanel, 1)
	drawLayerList(screen, sidebarRect, a.layer)

	drawPanel(screen, mapPanelRect, themePanel, 1)
	drawMapPanel(screen, mapPanelRect, a.snapshot, a.layer, a.selectedCell)

	drawPanel(screen, inspectorRect, themePanel, 1)
	drawInspector(screen, inspectorRect, a.snapshot, a.selectedCell)

	controlsRect := image.Rect(0, contentBottom, windowWidth, windowHeight)
	drawPanel(screen, controlsRect, themePanel, 1)
	drawControls(screen, controlsRect, a)
}

func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowWidth, windowHeight
}

func (a *App) reset() error {
	rebuilt, err := simulation.NewSimulationWithConfig(a.config)
	if err != nil {
		return err
	}
	a.simulation = rebuilt
	a.snapshot = rebuilt.Snapshot()
	a.playing = false
	a.yearsPerSecond = 1
	a.accumulatedStep = 0
	a.selectedCell = nil
	return nil
}

func (a *App) cycleSpeed(delta int) {
	if len(speedLevels) == 0 {
		return
	}
	current := a.displaySpeed()
	for i, v := range speedLevels {
		if v == current {
			idx := (i + delta + len(speedLevels)) % len(speedLevels)
			a.yearsPerSecond = speedLevels[idx]
			return
		}
	}
	a.yearsPerSecond = speedLevels[0]
}

func (a *App) displaySpeed() float64 {
	if a.yearsPerSecond <= 0 {
		return 1
	}
	return a.yearsPerSecond
}

func (a *App) mapViewport() image.Rectangle {
	panel := image.Rect(sidebarWidth, headerHeight, windowWidth-inspectorWidth, windowHeight-footerHeight)
	padding := 16
	top := panel.Min.Y + 52
	availableWidth := panel.Dx() - 2*padding
	availableHeight := panel.Max.Y - top - padding
	size := availableWidth
	if availableHeight < size {
		size = availableHeight
	}
	left := panel.Min.X + (panel.Dx()-size)/2
	return image.Rect(left, top, left+size, top+size)
}

func (a *App) mapCellFromMouse() (*CellSelection, bool) {
	posX, posY := ebiten.CursorPosition()
	viewport := a.mapViewport()
	if !pointInRect(posX, posY, viewport) {
		return nil, false
	}
	mapW, mapH := a.snapshot.TerrainMap.Width, a.snapshot.TerrainMap.Height
	cellW := float64(viewport.Dx()) / float64(mapW)
	cellH := float64(viewport.Dy()) / float64(mapH)
	cellX := int(float64(posX-viewport.Min.X) / cellW)
	cellY := int(float64(posY-viewport.Min.Y) / cellH)
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
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
	screen.DrawImage(img, opts)
	if border > 0 {
		drawBorderRect(screen, rect, themeBorder, border)
	}
}

func drawBorderRect(screen *ebiten.Image, rect image.Rectangle, col color.Color, thickness int) {
	if rect.Dx() <= 0 || rect.Dy() <= 0 || thickness <= 0 {
		return
	}
	for i := 0; i < thickness; i++ {
		line := ebiten.NewImage(rect.Dx(), 1)
		line.Fill(col)
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y+i))
		screen.DrawImage(line, opts)

		line2 := ebiten.NewImage(rect.Dx(), 1)
		line2.Fill(col)
		opts2 := &ebiten.DrawImageOptions{}
		opts2.GeoM.Translate(float64(rect.Min.X), float64(rect.Max.Y-1-i))
		screen.DrawImage(line2, opts2)

		colImg := ebiten.NewImage(1, rect.Dy())
		colImg.Fill(col)
		opts3 := &ebiten.DrawImageOptions{}
		opts3.GeoM.Translate(float64(rect.Min.X+i), float64(rect.Min.Y))
		screen.DrawImage(colImg, opts3)

		colImg2 := ebiten.NewImage(1, rect.Dy())
		colImg2.Fill(col)
		opts4 := &ebiten.DrawImageOptions{}
		opts4.GeoM.Translate(float64(rect.Max.X-1-i), float64(rect.Min.Y))
		screen.DrawImage(colImg2, opts4)
	}
}
