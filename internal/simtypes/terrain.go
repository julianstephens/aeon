package simtypes

const (
	DefaultMapWidth  = 65
	DefaultMapHeight = 65

	RoughnessDelta = 0.6
)

type Position struct {
	X, Y int
}

type LayerMap struct {
	Width  int
	Height int
	Values []float64
}

func NewLayerMap(width, height int) *LayerMap {
	return &LayerMap{
		Width:  width,
		Height: height,
		Values: make([]float64, width*height),
	}
}

func (em *LayerMap) Get(x, y int) float64 {
	return em.Values[y*em.Width+x]
}

func (em *LayerMap) Set(x, y int, value float64) {
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
}

type TerrainMap struct {
	initialized bool
	Width       int
	Height      int
	Cells       []TerrainCell
}

func NewTerrainMap(width, height int) TerrainMap {
	return TerrainMap{
		initialized: false,
		Width:       width,
		Height:      height,
		Cells:       make([]TerrainCell, width*height),
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
}

func (tm *TerrainMap) IsInitialized() bool {
	return tm.initialized
}

func (tm *TerrainMap) SetInitialized(initialized bool) {
	tm.initialized = initialized
}

// ApplyElevation applies the given LayerMap to the TerrainMap, updating the Elevation field of each TerrainCell.
func (tm *TerrainMap) ApplyElevation(layer *LayerMap) {
	for x := 0; x < tm.Width; x++ {
		for y := 0; y < tm.Height; y++ {
			elevation := layer.Get(x, y)
			cell := tm.GetCell(x, y)
			cell.Elevation = elevation
			tm.SetCell(x, y, *cell)
		}
	}
}

func (tm *TerrainMap) ApplyMoisture(layer *LayerMap) {
	for x := 0; x < tm.Width; x++ {
		for y := 0; y < tm.Height; y++ {
			moisture := layer.Get(x, y)
			cell := tm.GetCell(x, y)
			cell.Moisture = moisture
			tm.SetCell(x, y, *cell)
		}
	}
}
func (tm *TerrainMap) ApplyFertility(layer *LayerMap) {
	for x := 0; x < tm.Width; x++ {
		for y := 0; y < tm.Height; y++ {
			fertility := layer.Get(x, y)
			cell := tm.GetCell(x, y)
			cell.Fertility = fertility
			tm.SetCell(x, y, *cell)
		}
	}
}

func (tm *TerrainMap) ApplyTerrain(layer *LayerMap) {
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
