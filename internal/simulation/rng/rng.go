package rng

import (
	"encoding/binary"
	"math/rand/v2"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/mroth/weightedrand/v3"
)

const (
	// Agent attribute ranges
	MinAge = 1
	MaxAge = 100

	MinProductivity = 0.5
	MaxProductivity = 1.5

	MinHealth = 0
	MaxHealth = 100

	MinHealthResilience = 0.5
	MaxHealthResilience = 1.5

	MinFoodCapacity = 0
	MaxFoodCapacity = 100

	MinFertility = 0.5
	MaxFertility = 1.5

	MinWealth = 0
	MaxWealth = 1000

	// Terrain attribute ranges
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

type RNG struct {
	rnd *rand.Rand
}

func NewRNG(seed [32]byte) *RNG {
	return &RNG{
		rnd: rand.New(rand.NewPCG(derivePCG(seed))),
	}
}

type boundary struct {
	Min uint8
	Max uint8
}

var ageChooser = mustAgeChooser()

func mustAgeChooser() *weightedrand.Chooser[boundary, int] {
	chooser, err := weightedrand.NewChooser(
		weightedrand.NewChoice(boundary{Min: 1, Max: 14}, 25),  // ~25%
		weightedrand.NewChoice(boundary{Min: 15, Max: 29}, 25), // ~25%
		weightedrand.NewChoice(boundary{Min: 30, Max: 49}, 30), // ~30%
		weightedrand.NewChoice(boundary{Min: 50, Max: 69}, 15), // ~15%
		weightedrand.NewChoice(boundary{Min: 70, Max: 100}, 5), // ~5% (70+ capped at 100)
	)
	if err != nil {
		panic(err)
	}
	return chooser
}

// Age generates a random age between 1 and 100 (inclusive).
func (r *RNG) Age() uint8 {
	ageBoundary := ageChooser.PickWith(r.rnd)
	n := r.rnd.UintN(uint(ageBoundary.Max - ageBoundary.Min + 1)) //nolint:gosec // G115: safe, n ≤ 255
	return ageBoundary.Min + uint8(n)                             //nolint:gosec // G115: safe, n ≤ Max-Min
}

// AgeInRange generates a random age between minAge and maxAge (inclusive).
func (r *RNG) AgeInRange(minAge, maxAge uint8) uint8 {
	for {
		age := r.Age()
		if age >= minAge && age <= maxAge {
			return age
		}
	}
}

// Productivity generates a random productivity value between 0.5 and 1.5 (inclusive).
func (r *RNG) Productivity() float64 {
	return r.rnd.Float64()*(MaxProductivity-MinProductivity) + MinProductivity
}

// Health generates a random health value between 0 and 100 (inclusive).
func (r *RNG) Health() float64 {
	return r.rnd.Float64()*(MaxHealth-MinHealth) + MinHealth
}

// HealthInRange generates a random health value between minHealth and maxHealth (inclusive).
func (r *RNG) HealthInRange(minHealth, maxHealth float64) float64 {
	for {
		health := r.Health()
		if health >= minHealth && health <= maxHealth {
			return health
		}
	}
}

// HealthResilience generates a random health resilience value between 0.5 and 1.5 (inclusive).
func (r *RNG) HealthResilience() float64 {
	return r.rnd.Float64()*(MaxHealthResilience-MinHealthResilience) + MinHealthResilience
}

// FoodCapacity generates a random food capacity value between 0 and 100 (inclusive).
func (r *RNG) FoodCapacity() float64 {
	return r.rnd.Float64()*(MaxFoodCapacity-MinFoodCapacity) + MinFoodCapacity
}

// FoodCapacityInRange generates a random food capacity value between minFoodCapacity and maxFoodCapacity (inclusive).
func (r *RNG) FoodCapacityInRange(minFoodCapacity, maxFoodCapacity float64) float64 {
	for {
		foodCapacity := r.FoodCapacity()
		if foodCapacity >= minFoodCapacity && foodCapacity <= maxFoodCapacity {
			return foodCapacity
		}
	}
}

// Sex generates a random sex, either male (0) or female (1).
func (r *RNG) Sex() simtypes.Sex {
	return simtypes.Sex(r.rnd.IntN(2))
}

// Fertility generates a random fertility value between 0.5 and 1.5 (inclusive).
func (r *RNG) Fertility() float64 {
	return r.rnd.Float64()*(MaxFertility-MinFertility) + MinFertility
}

// Wealth generates a random wealth value between 0 and 1000 (inclusive).
func (r *RNG) Wealth() float64 {
	return r.rnd.Float64()*(MaxWealth-MinWealth) + MinWealth
}

// WealthInRange generates a random wealth value between minWealth and maxWealth (inclusive).
func (r *RNG) WealthInRange(minWealth, maxWealth float64) float64 {
	for {
		wealth := r.Wealth()
		if wealth >= minWealth && wealth <= maxWealth {
			return wealth
		}
	}
}

// Occupation generates a random occupation from a predefined list of occupations.
func (r *RNG) Occupation() simtypes.Occupation {
	return simtypes.Occupations[r.rnd.IntN(len(simtypes.Occupations))]
}

// Location generates a random location with X and Y coordinates between 0 and 63 (inclusive).
func (r *RNG) Location() simtypes.Position {
	return simtypes.Position{
		X: r.rnd.IntN(simtypes.DefaultMapWidth*simtypes.DefaultMapHeight) % simtypes.DefaultMapWidth,
		Y: r.rnd.IntN(simtypes.DefaultMapWidth*simtypes.DefaultMapHeight) % simtypes.DefaultMapHeight,
	}
}

// Elevation generates a random elevation value between 0 and 1 (inclusive).
// The roughness parameter can be used to influence the distribution of elevation values. If roughness is nil, a uniform distribution is used.
// If roughness is provided, a zero-mean normal distribution is used with the specified roughness as the standard deviation.
func (r *RNG) Elevation(roughness *float64) float64 {
	if roughness != nil {
		return r.rnd.NormFloat64() * (*roughness)
	}
	return r.rnd.Float64()
}

func derivePCG(digest [32]byte) (uint64, uint64) {
	chunk1 := digest[:16]
	chunk2 := digest[16:]

	return binary.LittleEndian.Uint64(chunk1), binary.LittleEndian.Uint64(chunk2)
}
