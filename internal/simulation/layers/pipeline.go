package layers

import (
	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

const (
	WaterThreshold    = 0.35
	MountainThreshold = 0.85
	ForestMoisture    = 0.60

	WaterTarget    = 0.35
	PlainsTarget   = 0.35
	ForestTarget   = 0.22
	MountainTarget = 0.08

	MinWaterRegionSize = 8

	MinSettlementDistance = 12

	PlainsFoodCapacity   = 100.0
	ForestFoodCapacity   = 75.0
	MountainFoodCapacity = 15.0
)

type Pipeline struct {
	tm         *simtypes.TerrainMap
	generator  *Generator
	classifier *TerrainClassifier
}

func NewPipeline(width, height int, baseRng *rng.RNG) *Pipeline {
	p := &Pipeline{
		tm:         simtypes.NewTerrainMap(width, height),
		generator:  NewGenerator(width, height, baseRng),
		classifier: NewTerrainClassifier(simtypes.NewTerrainMap(width, height)),
	}
	return p
}

// Run executes the terrain generation and classification pipeline, returning the final terrain map.
func (p *Pipeline) Run(worldSeed [32]byte) (*simtypes.TerrainMap, error) {
	generatorArtifacts, err := p.generator.GenerateLayers(worldSeed, p.tm)
	if err != nil {
		return nil, err
	}

	if err := p.classifier.Classify(
		generatorArtifacts.Elevation,
		generatorArtifacts.Moisture,
		generatorArtifacts.Fertility,
		generatorArtifacts.Terrain,
	); err != nil {
		return nil, err
	}

	return p.classifier.tm, nil
}
