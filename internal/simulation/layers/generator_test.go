package layers_test

import (
	"crypto/sha256"
	"math"
	"reflect"
	"sort"
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/layers"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

func TestGenerator_GenerateLayers_OnlyReturnsScalarLayers(t *testing.T) {
	worldSeed := seedFromString("scalar-only-artifacts")
	generator := layers.NewGenerator(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight, rng.NewRNG(worldSeed))

	artifacts, err := generator.GenerateIntrinsicLayers(worldSeed)
	if err != nil {
		t.Fatalf("GenerateLayers returned error: %v", err)
	}

	if artifacts.Elevation == nil || artifacts.Moisture == nil || artifacts.Fertility == nil {
		t.Fatal("expected all scalar layers to be generated")
	}

	artifactType := reflect.TypeOf(artifacts)
	if _, exists := artifactType.FieldByName("Terrain"); exists {
		t.Fatal("generator artifacts should not expose a discrete terrain layer")
	}
}

func TestGenerator_GenerateTerrainMap_IsDeterministicForSameSeed(t *testing.T) {
	seed := seedFromString("deterministic-terrain")

	tmA := generateTerrainMap(t, seed)
	tmB := generateTerrainMap(t, seed)

	if tmA.Width != tmB.Width || tmA.Height != tmB.Height {
		t.Fatalf("unexpected shape mismatch: (%d,%d) vs (%d,%d)", tmA.Width, tmA.Height, tmB.Width, tmB.Height)
	}

	for i := range tmA.Cells {
		a := tmA.Cells[i]
		b := tmB.Cells[i]

		if a.Elevation != b.Elevation {
			t.Fatalf("elevation mismatch at index %d: %f != %f", i, a.Elevation, b.Elevation)
		}
		if a.Moisture != b.Moisture {
			t.Fatalf("moisture mismatch at index %d: %f != %f", i, a.Moisture, b.Moisture)
		}
		if a.Fertility != b.Fertility {
			t.Fatalf("fertility mismatch at index %d: %f != %f", i, a.Fertility, b.Fertility)
		}
	}
}

func TestGenerator_GenerateTerrainMap_ChangesWithDifferentSeed(t *testing.T) {
	seedA := seedFromString("world-seed-a")
	seedB := seedFromString("world-seed-b")

	tmA := generateTerrainMap(t, seedA)
	tmB := generateTerrainMap(t, seedB)

	differentCells := 0
	for i := range tmA.Cells {
		a := tmA.Cells[i]
		b := tmB.Cells[i]
		if a.Elevation != b.Elevation || a.Moisture != b.Moisture || a.Fertility != b.Fertility {
			differentCells++
		}
	}

	if differentCells == 0 {
		t.Fatal("expected maps generated from different seeds to differ")
	}
}

func TestGenerator_GenerateTerrainMap_ProducesInitializedVariedAndSmoothLayers(t *testing.T) {
	tm := generateTerrainMap(t, seedFromString("quality-check"))

	if !tm.IsInitialized() {
		t.Fatal("terrain map should be marked initialized")
	}

	elevation := extractLayer(tm, func(c simtypes.TerrainCell) float64 { return c.Elevation })
	moisture := extractLayer(tm, func(c simtypes.TerrainCell) float64 { return c.Moisture })
	fertility := extractLayer(tm, func(c simtypes.TerrainCell) float64 { return c.Fertility })

	assertSpread(t, elevation, "elevation", 0.20)
	assertSpread(t, moisture, "moisture", 0.20)
	assertSpread(t, fertility, "fertility", 0.15)

	assertMeanNeighborDeltaBelow(t, tm, func(c simtypes.TerrainCell) float64 { return c.Elevation }, "elevation", 0.25)
	assertMeanNeighborDeltaBelow(t, tm, func(c simtypes.TerrainCell) float64 { return c.Moisture }, "moisture", 0.25)
	assertMeanNeighborDeltaBelow(t, tm, func(c simtypes.TerrainCell) float64 { return c.Fertility }, "fertility", 0.22)

	for i, v := range fertility {
		if v < 0.0 || v > 1.0 {
			t.Fatalf("fertility out of [0,1] at index %d: %f", i, v)
		}
	}
}

func TestGenerator_GenerateTerrainMap_FertilityTracksMoistureAndMidElevation(t *testing.T) {
	tm := generateTerrainMap(t, seedFromString("fertility-relationships"))

	moisture := extractLayer(tm, func(c simtypes.TerrainCell) float64 { return c.Moisture })
	fertility := extractLayer(tm, func(c simtypes.TerrainCell) float64 { return c.Fertility })
	elevation := extractLayer(tm, func(c simtypes.TerrainCell) float64 { return c.Elevation })

	meanLowMoistureFertility, meanHighMoistureFertility := quartileMeans(moisture, fertility)
	if !(meanHighMoistureFertility > meanLowMoistureFertility) {
		t.Fatalf(
			"expected higher moisture to increase fertility: low=%f high=%f",
			meanLowMoistureFertility,
			meanHighMoistureFertility,
		)
	}

	normElevation := normalizeSlice(elevation)
	extremeElevation := make([]float64, len(normElevation))
	for i, v := range normElevation {
		extremeElevation[i] = 2.0 * math.Abs(v-0.5)
	}

	fertilityVsExtremeElevation := correlation(extremeElevation, fertility)
	if !(fertilityVsExtremeElevation < 0.0) {
		t.Fatalf(
			"expected fertility to decrease as elevation becomes more extreme, corr=%f",
			fertilityVsExtremeElevation,
		)
	}

	fertilityVsMoisture := correlation(moisture, fertility)
	if !(fertilityVsMoisture > 0.20 && fertilityVsMoisture < 0.98) {
		t.Fatalf(
			"unexpected fertility-moisture correlation: got %f, want in (0.20, 0.98)",
			fertilityVsMoisture,
		)
	}
}

func generateTerrainMap(t *testing.T, worldSeed [32]byte) simtypes.TerrainMap {
	t.Helper()

	tm := simtypes.NewTerrainMap(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight)
	generator := layers.NewGenerator(tm.Width, tm.Height, rng.NewRNG(worldSeed))

	artifacts, err := generator.GenerateIntrinsicLayers(worldSeed)
	if err != nil {
		t.Fatalf("GenerateLayers returned error: %v", err)
	}

	tm.ApplyElevation(artifacts.Elevation)
	tm.ApplyMoisture(artifacts.Moisture)
	tm.ApplyFertility(artifacts.Fertility)
	tm.SetInitialized(true)

	return *tm
}

func extractLayer(tm simtypes.TerrainMap, project func(c simtypes.TerrainCell) float64) []float64 {
	values := make([]float64, 0, len(tm.Cells))
	for _, cell := range tm.Cells {
		values = append(values, project(cell))
	}
	return values
}

func assertSpread(t *testing.T, values []float64, label string, minSpread float64) {
	t.Helper()

	if len(values) == 0 {
		t.Fatalf("%s layer is empty", label)
	}

	p10 := percentile(values, 0.10)
	p90 := percentile(values, 0.90)
	spread := p90 - p10

	if spread < minSpread {
		t.Fatalf("%s spread too small: p10=%f p90=%f spread=%f (< %f)", label, p10, p90, spread, minSpread)
	}
}

func assertMeanNeighborDeltaBelow(
	t *testing.T,
	tm simtypes.TerrainMap,
	project func(c simtypes.TerrainCell) float64,
	label string,
	maxMeanDelta float64,
) {
	t.Helper()

	var sum float64
	count := 0

	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			center := tm.GetCell(x, y)
			if center == nil {
				continue
			}
			centerValue := project(*center)

			right := tm.GetCell(x+1, y)
			if right != nil {
				sum += math.Abs(centerValue - project(*right))
				count++
			}

			down := tm.GetCell(x, y+1)
			if down != nil {
				sum += math.Abs(centerValue - project(*down))
				count++
			}
		}
	}

	if count == 0 {
		t.Fatalf("%s has no neighbor comparisons", label)
	}

	meanDelta := sum / float64(count)
	if meanDelta > maxMeanDelta {
		t.Fatalf("%s appears too noisy: mean neighbor delta=%f (> %f)", label, meanDelta, maxMeanDelta)
	}
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	idx := int(math.Round(p * float64(len(sorted)-1)))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}

	return sorted[idx]
}

func quartileMeans(by, values []float64) (float64, float64) {
	if len(by) != len(values) || len(by) == 0 {
		return 0, 0
	}

	q25 := percentile(by, 0.25)
	q75 := percentile(by, 0.75)

	var lowSum, highSum float64
	lowCount, highCount := 0, 0

	for i := range by {
		if by[i] <= q25 {
			lowSum += values[i]
			lowCount++
		}
		if by[i] >= q75 {
			highSum += values[i]
			highCount++
		}
	}

	if lowCount == 0 || highCount == 0 {
		return 0, 0
	}

	return lowSum / float64(lowCount), highSum / float64(highCount)
}

func normalizeSlice(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}

	minValue := values[0]
	maxValue := values[0]
	for _, v := range values {
		if v < minValue {
			minValue = v
		}
		if v > maxValue {
			maxValue = v
		}
	}

	normalized := make([]float64, len(values))
	if maxValue == minValue {
		return normalized
	}

	for i, v := range values {
		normalized[i] = (v - minValue) / (maxValue - minValue)
	}

	return normalized
}

func correlation(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var sumA, sumB float64
	for i := range a {
		sumA += a[i]
		sumB += b[i]
	}

	meanA := sumA / float64(len(a))
	meanB := sumB / float64(len(b))

	var cov, varA, varB float64
	for i := range a {
		da := a[i] - meanA
		db := b[i] - meanB
		cov += da * db
		varA += da * da
		varB += db * db
	}

	if varA == 0 || varB == 0 {
		return 0
	}

	return cov / math.Sqrt(varA*varB)
}

func seedFromString(s string) [32]byte {
	return sha256.Sum256([]byte(s))
}
