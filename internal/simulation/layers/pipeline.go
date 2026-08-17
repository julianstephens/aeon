package layers

import (
	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

type Pipeline struct {
	tm         *simtypes.TerrainMap
	generator  *Generator
	classifier *TerrainClassifier
}

func NewPipeline(width, height int, baseRng *rng.RNG) *Pipeline {
	tm := simtypes.NewTerrainMap(width, height)
	return &Pipeline{
		tm:         tm,
		generator:  NewGenerator(width, height, baseRng),
		classifier: NewTerrainClassifier(tm),
	}
}

// Run executes terrain layer generation, applies the generated scalar layers,
// and classifies the final terrain map.
func (p *Pipeline) Run(worldSeed [32]byte) (*simtypes.TerrainMap, error) {
	artifacts, err := p.generator.GenerateLayers(worldSeed, p.tm)
	if err != nil {
		return nil, err
	}

	p.tm.ApplyElevation(artifacts.Elevation)
	p.tm.ApplyMoisture(artifacts.Moisture)
	p.tm.ApplyFertility(artifacts.Fertility)

	if err := p.classifier.Classify(
		artifacts.Elevation,
		artifacts.Moisture,
		artifacts.Fertility,
		artifacts.Terrain,
	); err != nil {
		return nil, err
	}

	return p.tm, nil
}
