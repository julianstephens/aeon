package layers

import (
	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

// DiamondSquareGenerator is a struct that generates terrain using the Diamond-Square algorithm.
type DSGenerator struct {
	size   int
	random *rng.RNG
}

func NewDSGenerator(size int, random *rng.RNG) *DSGenerator {
	return &DSGenerator{
		size:   size,
		random: random,
	}
}

// Generate creates a new LayerMap using the Diamond-Square algorithm.
// Based on https://janert.me/blog/2022/the-diamond-square-algorithm-for-terrain-generation/
// -> d = LayerMap, i/j = x/y coordinates, n = ds.size, w = ds.size-1 = step, v = (ds.size-1) // 2, s = roughness
func (ds *DSGenerator) Generate() *simtypes.Layer {
	m := &simtypes.Layer{
		Width:  ds.size,
		Height: ds.size,
		Values: make([]float64, ds.size*ds.size),
	}

	ds.initializeCorners(m)

	step := ds.size - 1
	roughness := 1.0

	for step > 1 {
		start := step / 2
		ds.diamondStep(m, start, step, roughness)
		ds.squareStep(m, start, step, roughness)
		step /= 2
		roughness *= simtypes.RoughnessDelta
	}

	return normalizeDSOutput(m)
}

func (ds *DSGenerator) initializeCorners(m *simtypes.Layer) {
	m.Set(0, 0, ds.random.Elevation(nil))
	m.Set(0, ds.size-1, ds.random.Elevation(nil))
	m.Set(ds.size-1, 0, ds.random.Elevation(nil))
	m.Set(ds.size-1, ds.size-1, ds.random.Elevation(nil))
}

func (ds *DSGenerator) diamondStep(m *simtypes.Layer, start int, step int, roughness float64) {
	diamond := [][]int{
		{-1, -1},
		{-1, 1},
		{1, 1},
		{1, -1},
	}

	for x := start; x < ds.size; x += step {
		for y := start; y < ds.size; y += step {
			m.Set(x, y, ds.getAverage(m, x, y, start, diamond)+ds.random.Elevation(&roughness))
		}
	}
}

func (ds *DSGenerator) squareStep(m *simtypes.Layer, start int, step int, roughness float64) {
	square := [][]int{
		{-1, 0},
		{0, -1},
		{1, 0},
		{0, 1},
	}

	// rows
	for x := start; x < ds.size; x += step {
		for y := 0; y < ds.size; y += step {
			m.Set(x, y, ds.getAverage(m, x, y, start, square)+ds.random.Elevation(&roughness))
		}
	}

	// columns
	for x := 0; x < ds.size; x += step {
		for y := start; y < ds.size; y += step {
			m.Set(x, y, ds.getAverage(m, x, y, start, square)+ds.random.Elevation(&roughness))
		}
	}
}

// Compute the average of the surrounding points for the diamond and square steps.
// Uses fixed boundary conditions, meaning that points outside the map are ignored in the average calculation.
func (ds *DSGenerator) getAverage(m *simtypes.Layer, x, y, step int, offsets [][]int) float64 {
	var sum float64
	var count int

	for _, offset := range offsets {
		p, q := offset[0], offset[1]
		pp, qq := x+p*step, y+q*step
		if 0 <= pp && pp < ds.size && 0 <= qq && qq < ds.size {
			sum += m.Get(pp, qq)
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func normalizeDSOutput(layer *simtypes.Layer) *simtypes.Layer {
	min, max := minMax(layer)
	normalizedLayer := simtypes.NewLayer(layer.Width, layer.Height)

	for x := 0; x < layer.Width; x++ {
		for y := 0; y < layer.Height; y++ {
			value := layer.Get(x, y)
			normalizedLayer.Set(x, y, clamp((value-min)/(max-min)))
		}
	}
	return normalizedLayer
}
