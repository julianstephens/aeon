package rng_test

import (
	"testing"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

func TestRNG_Age_IsDeterministic(t *testing.T) {
	assertDeterministic(t, 100, func(rng *rng.RNG) uint8 {
		return rng.Age()
	})
}

func TestRNG_AgeInRange_IsDeterministic(t *testing.T) {
	assertDeterministic(t, 100, func(rng *rng.RNG) uint8 {
		return rng.AgeInRange(18, 35)
	})
}

func TestRNG_Health_IsDeterministic(t *testing.T) {
	assertDeterministic(t, 100, func(rng *rng.RNG) float64 {
		return rng.Health()
	})
}

func TestRNG_HealthInRange_IsDeterministic(t *testing.T) {
	assertDeterministic(t, 100, func(rng *rng.RNG) float64 {
		return rng.HealthInRange(20.0, 80.0)
	})
}

func TestRNG_FoodCapacity_IsDeterministic(t *testing.T) {
	assertDeterministic(t, 100, func(rng *rng.RNG) float64 {
		return rng.FoodCapacity()
	})
}

func TestRNG_FoodCapacityInRange_IsDeterministic(t *testing.T) {
	assertDeterministic(t, 100, func(rng *rng.RNG) float64 {
		return rng.FoodCapacityInRange(10.0, 60.0)
	})
}

func TestRNG_Sex_IsDeterministic(t *testing.T) {
	assertDeterministic(t, 100, func(rng *rng.RNG) simtypes.Sex {
		return rng.Sex()
	})
}

func TestRNG_Wealth_IsDeterministic(t *testing.T) {
	assertDeterministic(t, 100, func(rng *rng.RNG) float64 {
		return rng.Wealth()
	})
}

func TestRNG_WealthInRange_IsDeterministic(t *testing.T) {
	assertDeterministic(t, 100, func(rng *rng.RNG) float64 {
		return rng.WealthInRange(100.0, 750.0)
	})
}

func TestRNG_Occupation_IsDeterministic(t *testing.T) {
	assertDeterministic(t, 100, func(rng *rng.RNG) simtypes.Occupation {
		return rng.Occupation()
	})
}

func TestRNG_Location_IsDeterministic(t *testing.T) {
	assertDeterministic(t, 100, func(rng *rng.RNG) simtypes.Position {
		return rng.Location()
	})
}

func TestDeriveSeed_IsStableForSamePath(t *testing.T) {
	seed := [32]byte{
		0x01, 0x0A, 0x14, 0x1E, 0x28, 0x32, 0x3C, 0x46,
		0x50, 0x5A, 0x64, 0x6E, 0x78, 0x82, 0x8C, 0x96,
		0xA0, 0xAA, 0xB4, 0xBE, 0xC8, 0xD2, 0xDC, 0xE6,
		0xF0, 0xFA, 0x04, 0x0E, 0x18, 0x22, 0x2C, 0x36,
	}

	retryA := rng.DeriveSeed(seed, "retry:1")
	retryB := rng.DeriveSeed(seed, "retry:1")

	if retryA != retryB {
		t.Fatal("expected retry derivation to be stable for identical input")
	}
}

func TestDeriveSeed_RetryPathsAreDistinct(t *testing.T) {
	seed := [32]byte{
		0x01, 0x0A, 0x14, 0x1E, 0x28, 0x32, 0x3C, 0x46,
		0x50, 0x5A, 0x64, 0x6E, 0x78, 0x82, 0x8C, 0x96,
		0xA0, 0xAA, 0xB4, 0xBE, 0xC8, 0xD2, 0xDC, 0xE6,
		0xF0, 0xFA, 0x04, 0x0E, 0x18, 0x22, 0x2C, 0x36,
	}

	retry0 := rng.DeriveSeed(seed, "retry:0")
	retry1 := rng.DeriveSeed(seed, "retry:1")
	retry2 := rng.DeriveSeed(seed, "retry:2")

	if retry0 == retry1 || retry1 == retry2 || retry0 == retry2 {
		t.Fatal("expected retry paths to produce distinct seeds")
	}
}

func assertDeterministic[T comparable](t *testing.T, calls int, sample func(rng *rng.RNG) T) {
	t.Helper()

	seed := [32]byte{
		0x01, 0x0A, 0x14, 0x1E, 0x28, 0x32, 0x3C, 0x46,
		0x50, 0x5A, 0x64, 0x6E, 0x78, 0x82, 0x8C, 0x96,
		0xA0, 0xAA, 0xB4, 0xBE, 0xC8, 0xD2, 0xDC, 0xE6,
		0xF0, 0xFA, 0x04, 0x0E, 0x18, 0x22, 0x2C, 0x36,
	}

	rngA := rng.NewRNG(seed)
	rngB := rng.NewRNG(seed)

	for i := range calls {
		valueA := sample(rngA)
		valueB := sample(rngB)

		if valueA != valueB {
			t.Fatalf("non-deterministic output on draw %d: got %v and %v", i+1, valueA, valueB)
		}
	}
}
