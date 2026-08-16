package simtypes

const (
	DefaultMapWidth  = 65
	DefaultMapHeight = 65

	RoughnessDelta = 0.6
)

type Position struct {
	X, Y int
}

type ElevationMap struct {
	Width  int
	Height int
	Values []float64
}

func NewElevationMap(width, height int) *ElevationMap {
	return &ElevationMap{
		Width:  width,
		Height: height,
		Values: make([]float64, width*height),
	}
}

func (em *ElevationMap) Get(x, y int) float64 {
	return em.Values[y*em.Width+x]
}

func (em *ElevationMap) Set(x, y int, elevation float64) {
	em.Values[y*em.Width+x] = elevation
}

type TerrainType int

const (
	TerrainTypePlains TerrainType = iota
	TerrainTypeForest
	TerrainTypeMountain // impassible
	TerrainTypeWater    // impassible
)

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

// ApplyElevation applies the given ElevationMap to the TerrainMap, updating the Elevation field of each TerrainCell.
func (tm *TerrainMap) ApplyElevation(layer ElevationMap) {
	for x := 0; x < tm.Width; x++ {
		for y := 0; y < tm.Height; y++ {
			elevation := layer.Get(x, y)
			cell := tm.GetCell(x, y)
			cell.Elevation = elevation
			tm.SetCell(x, y, *cell)
		}
	}
}

// func (tm *TerrainMap) ApplyMoisture(layer MoistureMap)
// func (tm *TerrainMap) ApplyFertility(layer FertilityMap)
