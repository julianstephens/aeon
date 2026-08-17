package layers

import (
	"fmt"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

const MaxGenerationAttempts = 10

type Pipeline struct {
	width      int
	height     int
	generator  *Generator
	classifier *TerrainClassifier
	validator  *WorldValidator
}

func NewPipeline(width, height int, baseRng *rng.RNG) *Pipeline {
	return &Pipeline{
		width:     width,
		height:    height,
		generator: NewGenerator(width, height, baseRng),
		validator: NewWorldValidator(DefaultViabilityRules()),
	}
}

// Run executes terrain layer generation, applies the generated scalar layers,
// and classifies the final terrain map.
func (p *Pipeline) Run(worldSeed [32]byte) (*simtypes.TerrainMap, error) {
	tm := simtypes.NewTerrainMap(p.width, p.height)
	p.classifier = NewTerrainClassifier(tm)

	artifacts, err := p.generator.GenerateIntrinsicLayers(worldSeed)
	if err != nil {
		return nil, err
	}

	tm.ApplyElevation(artifacts.Elevation)
	tm.ApplyMoisture(artifacts.Moisture)
	tm.ApplyFertility(artifacts.Fertility)

	if err := p.classifier.Classify(
		artifacts.Elevation,
		artifacts.Moisture,
		artifacts.Fertility,
	); err != nil {
		return nil, err
	}

	if _, err := p.validator.Validate(*tm); err != nil {
		return nil, &PipelineError{
			Code:    CodeGenerationError,
			Message: "Generated world failed viability validation",
			Cause:   err,
		}
	}

	tm.ApplyFoodCapacity(
		p.generator.GenerateComputedLayers(tm, artifacts.Elevation, artifacts.Moisture, artifacts.Fertility),
	)

	return tm, nil
}

// GenerateWorld retries generation with deterministic retry seeds until it
// finds a viable world or exhausts all attempts.
func (p *Pipeline) GenerateWorld(worldSeed [32]byte) (*simtypes.TerrainMap, error) {
	var lastErr error

	for attempt := 0; attempt < MaxGenerationAttempts; attempt++ {
		retrySeed := rng.DeriveSeed(worldSeed, fmt.Sprintf("retry:%d", attempt))
		tm, err := p.Run(retrySeed)
		if err == nil {
			return tm, nil
		}
		lastErr = err
	}

	return nil, &PipelineError{
		Code:    CodeGenerationError,
		Message: fmt.Sprintf("Failed to generate viable world after %d attempts", MaxGenerationAttempts),
		Cause:   lastErr,
	}
}
