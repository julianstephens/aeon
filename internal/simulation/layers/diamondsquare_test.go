package layers

import (
	"math"
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

func TestDSGenerator_Generate_HasExpectedShape(t *testing.T) {
	ds := NewDSGenerator(65, rng.NewRNG(testSeed()))
	em := ds.Generate()

	if em.Width != 65 {
		t.Fatalf("unexpected width: got %d, want %d", em.Width, 65)
	}

	if em.Height != 65 {
		t.Fatalf("unexpected height: got %d, want %d", em.Height, 65)
	}

	if len(em.Values) != 65*65 {
		t.Fatalf("unexpected values length: got %d, want %d", len(em.Values), 65*65)
	}
}

func TestDSGenerator_Generate_IsDeterministicForSameSeed(t *testing.T) {
	seed := testSeed()

	dsA := NewDSGenerator(65, rng.NewRNG(seed))
	dsB := NewDSGenerator(65, rng.NewRNG(seed))

	emA := dsA.Generate()
	emB := dsB.Generate()

	if len(emA.Values) != len(emB.Values) {
		t.Fatalf("mismatched value lengths: %d vs %d", len(emA.Values), len(emB.Values))
	}

	for i := range emA.Values {
		if emA.Values[i] != emB.Values[i] {
			t.Fatalf("non-deterministic output at index %d: %f != %f", i, emA.Values[i], emB.Values[i])
		}
	}
}

func TestDSGenerator_Generate_ProducesVariation(t *testing.T) {
	ds := NewDSGenerator(65, rng.NewRNG(testSeed()))
	em := ds.Generate()

	if len(em.Values) == 0 {
		t.Fatal("expected non-empty elevation values")
	}

	minValue := math.Inf(1)
	maxValue := math.Inf(-1)

	for _, v := range em.Values {
		if v < minValue {
			minValue = v
		}
		if v > maxValue {
			maxValue = v
		}
	}

	if !(maxValue > minValue) {
		t.Fatalf("expected variation in elevation values, got min=%f max=%f", minValue, maxValue)
	}
}

func TestDSGenerator_InitializeCorners_SetsCornersInUnitInterval(t *testing.T) {
	ds := NewDSGenerator(9, rng.NewRNG(testSeed()))
	em := simtypes.NewLayer(9, 9)

	ds.initializeCorners(em)

	assertInUnitInterval(t, em.Get(0, 0), "top-left")
	assertInUnitInterval(t, em.Get(0, 8), "bottom-left")
	assertInUnitInterval(t, em.Get(8, 0), "top-right")
	assertInUnitInterval(t, em.Get(8, 8), "bottom-right")
}

func TestDSGenerator_GetAverage_IgnoresOutOfBoundsNeighbors(t *testing.T) {
	ds := NewDSGenerator(3, rng.NewRNG(testSeed()))
	em := simtypes.NewLayer(3, 3)

	em.Set(0, 1, 0.2)
	em.Set(1, 0, 0.4)
	em.Set(2, 1, 0.8)
	em.Set(1, 2, 1.0)

	offsets := [][]int{
		{-1, 0},
		{0, -1},
		{1, 0},
		{0, 1},
	}

	averageAtCorner := ds.getAverage(em, 0, 0, 1, offsets)
	expectedAtCorner := (0.2 + 0.4) / 2.0
	if !nearlyEqual(averageAtCorner, expectedAtCorner, 1e-12) {
		t.Fatalf("unexpected corner average: got %f, want %f", averageAtCorner, expectedAtCorner)
	}

	averageAtCenter := ds.getAverage(em, 1, 1, 1, offsets)
	expectedAtCenter := (0.2 + 0.4 + 0.8 + 1.0) / 4.0
	if !nearlyEqual(averageAtCenter, expectedAtCenter, 1e-12) {
		t.Fatalf("unexpected center average: got %f, want %f", averageAtCenter, expectedAtCenter)
	}
}

func assertInUnitInterval(t *testing.T, value float64, label string) {
	t.Helper()

	if value < 0 || value > 1 {
		t.Fatalf("%s corner out of [0,1]: %f", label, value)
	}
}

func nearlyEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) <= epsilon
}

func testSeed() [32]byte {
	return [32]byte{
		0x01, 0x0A, 0x14, 0x1E, 0x28, 0x32, 0x3C, 0x46,
		0x50, 0x5A, 0x64, 0x6E, 0x78, 0x82, 0x8C, 0x96,
		0xA0, 0xAA, 0xB4, 0xBE, 0xC8, 0xD2, 0xDC, 0xE6,
		0xF0, 0xFA, 0x04, 0x0E, 0x18, 0x22, 0x2C, 0x36,
	}
}
