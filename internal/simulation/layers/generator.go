package layers

import (
	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

type Generator struct {
	width, height int
	random        *rng.RNG
	dsGenerator   *DSGenerator
}

func NewGenerator(width, height int, randomSource *rng.RNG) *Generator {
	return &Generator{
		width:       width,
		height:      height,
		random:      randomSource,
		dsGenerator: NewDSGenerator(max(width, height), randomSource),
	}
}

func (g *Generator) GenerateTerrainMap(tm *simtypes.TerrainMap) error {
	// generate elevation map using Diamond-Square algorithm
	elevationMap := g.dsGenerator.Generate()
	tm.ApplyElevation(*elevationMap)
	tm.SetInitialized(true)
	return nil
}
