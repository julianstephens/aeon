package simulation

import (
	"fmt"
	"math"

	"github.com/julianstephens/aeon/internal/simtypes"
)

type GUIConfig struct {
	Seed              string
	InitialPopulation float64
	GrowthRate        float64
	StarvationRate    float64
	MigrationRate     float64
	MaxCellCapacity   float64
}

func DefaultGUIConfig() GUIConfig {
	return GUIConfig{
		Seed:              DefaultExperimentSeed,
		InitialPopulation: DefaultExperimentInitialPopulation,
		GrowthRate:        DefaultExperimentGrowthRate,
		StarvationRate:    DefaultExperimentStarvationRate,
		MigrationRate:     DefaultExperimentMigrationRate,
		MaxCellCapacity:   DefaultExperimentMaxCellCapacity,
	}
}

type TerrainSnapshot struct {
	Width     int
	Height    int
	Terrain   []simtypes.TerrainType
	Elevation []float64
	Moisture  []float64
	Fertility []float64
	Food      []float64
}

type PopulationSnapshot struct {
	Population       []float64
	CarryingCapacity []float64
	TotalPopulation  float64
	OccupiedCells    int
	Utilization      float64
}

type SimulationSnapshot struct {
	Year       int
	TerrainMap TerrainSnapshot
	Population PopulationSnapshot
}

type SimulationStep struct {
	Year               int
	MigratedPopulation float64
	SourceCells        int
	DestinationCells   int
	AverageDistance    float64
	MaxDistance        int
	StarvationDeaths   float64
}

type Simulation struct {
	world    *World
	snapshot SimulationSnapshot
	params   PopulationParameters
	seed     string
	lastStep SimulationStep
}

func NewSimulation(seed string, params PopulationParameters) (*Simulation, error) {
	world, err := NewWorld(seed, params)
	if err != nil {
		return nil, err
	}
	if world == nil {
		return nil, fmt.Errorf("simulation world is nil")
	}

	sim := &Simulation{
		world:  world,
		params: params,
		seed:   seed,
	}
	sim.refreshSnapshot()
	return sim, nil
}

func NewSimulationWithConfig(cfg GUIConfig) (*Simulation, error) {
	params := PopulationParameters{
		GrowthRate:      cfg.GrowthRate,
		StarvationRate:  cfg.StarvationRate,
		MigrationRate:   cfg.MigrationRate,
		MaxCellCapacity: cfg.MaxCellCapacity,
	}
	world, err := NewWorld(cfg.Seed, params)
	if err != nil {
		return nil, err
	}
	if world == nil {
		return nil, fmt.Errorf("simulation world is nil")
	}
	if cfg.InitialPopulation > 0 {
		seeded := world.populationModel.SeedPopulation(world.terrainMap, cfg.InitialPopulation)
		if seeded == 0 && cfg.InitialPopulation > 0 {
			return nil, fmt.Errorf("failed to seed requested initial population")
		}
	}

	sim := &Simulation{
		world:  world,
		params: params,
		seed:   cfg.Seed,
	}
	sim.refreshSnapshot()
	return sim, nil
}

func (s *Simulation) World() *World {
	if s == nil {
		return nil
	}
	return s.world
}

func (s *Simulation) Step() SimulationStep {
	if s == nil || s.world == nil || s.world.populationModel == nil || s.world.terrainMap == nil {
		return SimulationStep{}
	}

	result := s.world.populationModel.AdvanceOneYear(s.world.terrainMap)
	s.lastStep = SimulationStep{
		Year:               s.world.Year(),
		MigratedPopulation: result.MigratedPopulation,
		SourceCells:        result.SourceCells,
		DestinationCells:   result.DestinationCells,
		AverageDistance:    result.AverageDistance,
		MaxDistance:        result.MaxDistance,
		StarvationDeaths:   result.StarvationDeaths,
	}
	s.refreshSnapshot()
	return s.lastStep
}

func (s *Simulation) Snapshot() SimulationSnapshot {
	if s == nil {
		return SimulationSnapshot{}
	}
	if s.world == nil {
		return s.snapshot
	}
	s.refreshSnapshot()
	return s.snapshot
}

func (s *Simulation) refreshSnapshot() {
	if s == nil || s.world == nil || s.world.terrainMap == nil {
		return
	}

	tm := s.world.terrainMap
	terrain := make([]simtypes.TerrainType, 0, len(tm.Cells))
	elevation := make([]float64, 0, len(tm.Cells))
	moisture := make([]float64, 0, len(tm.Cells))
	fertility := make([]float64, 0, len(tm.Cells))
	food := make([]float64, 0, len(tm.Cells))
	population := make([]float64, 0, len(tm.Cells))
	carryingCapacity := make([]float64, 0, len(tm.Cells))

	var totalPopulation float64
	occupiedCells := 0
	var totalCapacity float64
	for i := range tm.Cells {
		cell := &tm.Cells[i]
		terrain = append(terrain, cell.Terrain)
		elevation = append(elevation, cell.Elevation)
		moisture = append(moisture, cell.Moisture)
		fertility = append(fertility, cell.Fertility)
		food = append(food, cell.FoodCapacity)
		population = append(population, cell.Population)
		cap := s.world.populationModel.carryingCapacity(cell)
		carryingCapacity = append(carryingCapacity, cap)
		totalPopulation += cell.Population
		if cell.Population > 0 {
			occupiedCells++
		}
		totalCapacity += cap
	}

	utilization := 0.0
	if totalCapacity > 0 {
		utilization = totalPopulation / totalCapacity
	}

	s.snapshot = SimulationSnapshot{
		Year: s.world.Year(),
		TerrainMap: TerrainSnapshot{
			Width:     tm.Width,
			Height:    tm.Height,
			Terrain:   terrain,
			Elevation: elevation,
			Moisture:  moisture,
			Fertility: fertility,
			Food:      food,
		},
		Population: PopulationSnapshot{
			Population:       population,
			CarryingCapacity: carryingCapacity,
			TotalPopulation:  totalPopulation,
			OccupiedCells:    occupiedCells,
			Utilization:      utilization,
		},
	}
}

func (s *Simulation) Seed(initialPopulation float64) error {
	if s == nil || s.world == nil || s.world.terrainMap == nil || s.world.populationModel == nil {
		return fmt.Errorf("simulation is not initialized")
	}
	if initialPopulation < 0 {
		return fmt.Errorf("initial population must be >= 0")
	}
	seeded := s.world.populationModel.SeedPopulation(s.world.terrainMap, initialPopulation)
	if seeded == 0 && initialPopulation > 0 {
		return fmt.Errorf("failed to seed requested initial population")
	}
	s.refreshSnapshot()
	return nil
}

func (s *Simulation) Rebuild(seed string) error {
	if s == nil {
		return fmt.Errorf("simulation is nil")
	}
	newSim, err := NewSimulation(seed, s.params)
	if err != nil {
		return err
	}
	*s = *newSim
	return nil
}

func normalizeLayer(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	out := make([]float64, len(values))
	min := values[0]
	max := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if math.Abs(max-min) < 1e-12 {
		for i := range out {
			out[i] = 0.5
		}
		return out
	}
	for i, v := range values {
		out[i] = (v - min) / (max - min)
	}
	return out
}
