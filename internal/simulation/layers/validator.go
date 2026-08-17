package layers

import (
	"fmt"
	"strings"

	"github.com/julianstephens/aeon/internal/simtypes"
)

type ViabilityRules struct {
	MinPassableLand      float64
	MaxWater             float64
	MaxMountain          float64
	MinLargestLandRegion int
}

type WorldValidator struct {
	rules ViabilityRules
}

type WorldValidationError struct {
	Reasons []string
}

func (e *WorldValidationError) Error() string {
	return strings.Join(e.Reasons, "; ")
}

func DefaultViabilityRules() ViabilityRules {
	return ViabilityRules{
		MinPassableLand:      0.25,
		MaxWater:             0.60,
		MaxMountain:          0.20,
		MinLargestLandRegion: 100,
	}
}

func NewWorldValidator(rules ViabilityRules) *WorldValidator {
	return &WorldValidator{rules: rules}
}

func (v *WorldValidator) Rules() ViabilityRules {
	return v.rules
}

func (v *WorldValidator) ValidateFromDiagnostics(diagnostics WorldDiagnostics) error {
	return ValidateWorld(diagnostics, v.rules)
}

func (v *WorldValidator) Validate(tm simtypes.TerrainMap) (WorldDiagnostics, error) {
	diagnostics := ComputeWorldDiagnostics(tm)
	return diagnostics, ValidateWorld(diagnostics, v.rules)
}

func ValidateWorld(diagnostics WorldDiagnostics, rules ViabilityRules) error {
	reasons := make([]string, 0)

	if diagnostics.PassableLandPercentage < rules.MinPassableLand {
		reasons = append(
			reasons,
			fmt.Sprintf(
				"passable land %.1f%% below minimum %.1f%%",
				diagnostics.PassableLandPercentage*100,
				rules.MinPassableLand*100,
			),
		)
	}

	waterPercentage := diagnostics.TerrainPercentages[simtypes.TerrainTypeWater]
	if waterPercentage > rules.MaxWater {
		reasons = append(
			reasons,
			fmt.Sprintf("water %.1f%% exceeds maximum %.1f%%", waterPercentage*100, rules.MaxWater*100),
		)
	}

	mountainPercentage := diagnostics.TerrainPercentages[simtypes.TerrainTypeMountain]
	if mountainPercentage > rules.MaxMountain {
		reasons = append(
			reasons,
			fmt.Sprintf("mountain %.1f%% exceeds maximum %.1f%%", mountainPercentage*100, rules.MaxMountain*100),
		)
	}

	if diagnostics.LargestPassableLandRegion < rules.MinLargestLandRegion {
		reasons = append(
			reasons,
			fmt.Sprintf(
				"largest land region %d below minimum %d",
				diagnostics.LargestPassableLandRegion,
				rules.MinLargestLandRegion,
			),
		)
	}

	if len(reasons) == 0 {
		return nil
	}

	return &WorldValidationError{Reasons: reasons}
}
