package layers

import (
	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/rng"
	"github.com/julianstephens/go-utils/logger"
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
	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("generating elevation map")
	elevationMap := g.dsGenerator.Generate()
	logger.WithFields(map[string]interface{}{
		"width":  elevationMap.Width,
		"height": elevationMap.Height,
	}).Debug("elevation map generated")

	logger.WithFields(map[string]interface{}{
		"width":  g.width,
		"height": g.height,
	}).Debug("applying elevation map to terrain map")
	tm.ApplyElevation(*elevationMap)
	logger.Debug("elevation map applied to terrain map")
	tm.SetInitialized(true)
	return nil
}
