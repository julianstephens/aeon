package simtypes

import "errors"

const (
	DefaultMapWidth  = 65
	DefaultMapHeight = 65

	RoughnessDelta = 0.6
)

type Position struct {
	X, Y int
}

type Layer struct {
	Width  int
	Height int
	Values []float64
}

func NewLayer(width, height int) *Layer {
	return &Layer{
		Width:  width,
		Height: height,
		Values: make([]float64, width*height),
	}
}

func (em *Layer) Get(x, y int) float64 {
	return em.Values[y*em.Width+x]
}

func (em *Layer) Set(x, y int, value float64) {
	em.Values[y*em.Width+x] = value
}

type TerrainType int

const (
	TerrainTypePlains TerrainType = iota
	TerrainTypeForest
	TerrainTypeMountain // impassible
	TerrainTypeWater    // impassible
)

func isValidTerrainType(tt TerrainType) bool {
	switch tt {
	case TerrainTypePlains, TerrainTypeForest, TerrainTypeMountain, TerrainTypeWater:
		return true
	default:
		return false
	}
}

func (tt TerrainType) IsPassable() bool {
	switch tt {
	case TerrainTypePlains, TerrainTypeForest:
		return true
	case TerrainTypeMountain, TerrainTypeWater:
		return false
	default:
		return false
	}
}

func (tt TerrainType) String() string {
	switch tt {
	case TerrainTypePlains:
		return "Plains"
	case TerrainTypeForest:
		return "Forest"
	case TerrainTypeMountain:
		return "Mountain"
	case TerrainTypeWater:
		return "Water"
	default:
		return "Unknown"
	}
}

type TerrainCell struct {
	Location     Position
	Terrain      TerrainType
	Elevation    float64
	Moisture     float64
	Fertility    float64
	FoodCapacity float64
	Population   float64
}

type TerrainMap struct {
	initialized bool
	Width       int
	Height      int
	Cells       []TerrainCell
}

func NewTerrainMap(width, height int) *TerrainMap {
	cells := make([]TerrainCell, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cells[y*width+x] = TerrainCell{Location: Position{X: x, Y: y}}
		}
	}
	return &TerrainMap{
		initialized: false,
		Width:       width,
		Height:      height,
		Cells:       cells,
	}
}

func (tm *TerrainMap) GetCell(x, y int) *TerrainCell {
	if x < 0 || x >= tm.Width || y < 0 || y >= tm.Height {
		return nil
	}
	return &tm.Cells[y*tm.Width+x]
}

func (tm *TerrainMap) SetCell(x, y int, cell TerrainCell) {
	if x < 0 || x >= tm.Width || y < 0 || y >= tm.Height {
		return
	}
	tm.Cells[y*tm.Width+x] = cell
	tm.initialized = true
}

func (tm *TerrainMap) SetTerrainType(x, y int, terrainType TerrainType) error {
	cell := tm.GetCell(x, y)
	if cell != nil {
		cell.Terrain = terrainType
		tm.SetCell(x, y, *cell)
		return nil
	}
	return errors.New("invalid coordinates for terrain map")
}

func (tm *TerrainMap) IsInitialized() bool {
	return tm.initialized
}

func (tm *TerrainMap) SetInitialized(initialized bool) {
	tm.initialized = initialized
}

// ApplyElevation applies the given Layer to the TerrainMap, updating the Elevation field of each TerrainCell.
func (tm *TerrainMap) ApplyElevation(layer *Layer) {
	for x := 0; x < tm.Width; x++ {
		for y := 0; y < tm.Height; y++ {
			elevation := layer.Get(x, y)
			cell := tm.GetCell(x, y)
			cell.Elevation = elevation
			tm.SetCell(x, y, *cell)
		}
	}
}

func (tm *TerrainMap) ApplyMoisture(layer *Layer) {
	for x := 0; x < tm.Width; x++ {
		for y := 0; y < tm.Height; y++ {
			moisture := layer.Get(x, y)
			cell := tm.GetCell(x, y)
			cell.Moisture = moisture
			tm.SetCell(x, y, *cell)
		}
	}
}

func (tm *TerrainMap) ApplyFertility(layer *Layer) {
	for x := 0; x < tm.Width; x++ {
		for y := 0; y < tm.Height; y++ {
			fertility := layer.Get(x, y)
			cell := tm.GetCell(x, y)
			cell.Fertility = fertility
			tm.SetCell(x, y, *cell)
		}
	}
}

func (tm *TerrainMap) ApplyTerrain(layer *Layer) {
	for x := 0; x < tm.Width; x++ {
		for y := 0; y < tm.Height; y++ {
			terrainValue := layer.Get(x, y)
			cell := tm.GetCell(x, y)
			terrainType := TerrainType(int(terrainValue))
			if !isValidTerrainType(terrainType) {
				terrainType = TerrainTypePlains // default to plains if invalid
			}
			cell.Terrain = terrainType
			tm.SetCell(x, y, *cell)
		}
	}
}

func (tm *TerrainMap) ApplyFoodCapacity(layer *Layer) {
	for x := 0; x < tm.Width; x++ {
		for y := 0; y < tm.Height; y++ {
			foodCapacity := layer.Get(x, y)
			cell := tm.GetCell(x, y)
			cell.FoodCapacity = foodCapacity
			tm.SetCell(x, y, *cell)
		}
	}
}
