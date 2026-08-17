package simulation

import (
	"crypto/sha256"
	"fmt"
	"math"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/layers"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

const (
	DefaultExperimentSeed              = "default_world_001"
	DefaultExperimentYears             = 100
	DefaultExperimentInitialPopulation = 1000
	DefaultExperimentGrowthRate        = 0.025
	DefaultExperimentStarvationRate    = 0.10
	DefaultExperimentMigrationRate     = 0.07
	DefaultExperimentMaxCellCapacity   = 30
	DefaultExperimentInterval          = 25
)

type ExperimentConfig struct {
	Seed              string
	Years             int
	InitialPopulation float64
	GrowthRate        float64
	StarvationRate    float64
	MigrationRate     float64
	MaxCellCapacity   float64
	Interval          int
}

type Experiment struct {
	config ExperimentConfig
}

type ExperimentResult struct {
	Config  ExperimentConfig   `json:"config"`
	Samples []PopulationSample `json:"samples"`
}

type PopulationSample struct {
	Year              int                       `json:"year"`
	Population        float64                   `json:"population"`
	CarryingCapacity  float64                   `json:"carrying_capacity"`
	Utilization       float64                   `json:"utilization"`
	OccupiedCells     int                       `json:"occupied_cells"`
	OverCapacityCells int                       `json:"over_capacity_cells"`
	ByTerrain         []TerrainPopulationSample `json:"population_by_terrain"`

	MigratedPopulation float64 `json:"migrated_population"`
	SourceCells        int     `json:"source_cells"`
	DestinationCells   int     `json:"destination_cells"`
	AverageDistance    float64 `json:"average_distance"`
	MaxDistance        int     `json:"max_distance"`

	StarvationDeaths float64 `json:"starvation_deaths"`
}

type TerrainPopulationSample struct {
	Terrain          string  `json:"terrain"`
	Population       float64 `json:"population"`
	CarryingCapacity float64 `json:"carrying_capacity"`
	Utilization      float64 `json:"utilization"`
	OccupiedCapacity float64 `json:"occupied_capacity"`
	OccupiedRatio    float64 `json:"occupied_ratio"`
}

type intervalAccumulator struct {
	migratedPopulation float64
	sourceCells        int
	destinationCells   int
	distanceWeight     float64
	maxDistance        int
	starvationDeaths   float64
}

func DefaultExperimentConfig() ExperimentConfig {
	return ExperimentConfig{
		Seed:              DefaultExperimentSeed,
		Years:             DefaultExperimentYears,
		InitialPopulation: DefaultExperimentInitialPopulation,
		GrowthRate:        DefaultExperimentGrowthRate,
		StarvationRate:    DefaultExperimentStarvationRate,
		MigrationRate:     DefaultExperimentMigrationRate,
		MaxCellCapacity:   DefaultExperimentMaxCellCapacity,
		Interval:          DefaultExperimentInterval,
	}
}

func ValidateExperimentConfig(config ExperimentConfig) error {
	switch {
	case config.Years < 1:
		return fmt.Errorf("years must be at least 1")
	case config.InitialPopulation < 0:
		return fmt.Errorf("population must be greater than or equal to 0")
	case config.GrowthRate < 0:
		return fmt.Errorf("growth-rate must be greater than or equal to 0")
	case config.StarvationRate < 0:
		return fmt.Errorf("starvation-rate must be greater than or equal to 0")
	case config.MigrationRate < 0:
		return fmt.Errorf("migration-rate must be greater than or equal to 0")
	case config.MigrationRate > 1:
		return fmt.Errorf("migration-rate must be less than or equal to 1")
	case config.MaxCellCapacity <= 0:
		return fmt.Errorf("max-cell-capacity must be greater than 0")
	case config.Interval < 1:
		return fmt.Errorf("interval must be at least 1")
	case config.Interval > config.Years:
		return fmt.Errorf("interval must be less than or equal to years")
	default:
		return nil
	}
}

func NewExperiment(config ExperimentConfig) *Experiment {
	return &Experiment{config: config}
}

func (e *Experiment) Run() ExperimentResult {
	result, _ := e.RunE()
	return result
}

func (e *Experiment) RunE() (ExperimentResult, error) {
	result := ExperimentResult{Config: e.config, Samples: []PopulationSample{}}
	if err := ValidateExperimentConfig(e.config); err != nil {
		return result, err
	}

	terrainMap, err := generateExperimentWorld(e.config.Seed)
	if err != nil {
		return result, err
	}

	populationModel := NewPopulationModel(PopulationParameters{
		GrowthRate:      e.config.GrowthRate,
		StarvationRate:  e.config.StarvationRate,
		MigrationRate:   e.config.MigrationRate,
		MaxCellCapacity: e.config.MaxCellCapacity,
	})
	requestedInitialPopulation := int(math.Round(e.config.InitialPopulation))
	seededPopulation := populationModel.SeedPopulation(terrainMap, e.config.InitialPopulation)
	if e.config.MigrationRate == 0 {
		seededPopulation = populationModel.SeedPopulationBroadly(terrainMap, e.config.InitialPopulation)
	}
	if seededPopulation != requestedInitialPopulation {
		metrics := CollectPopulationMetrics(terrainMap, populationModel)
		return result, fmt.Errorf(
			"failed to seed requested initial population: requested %d, seeded %d (estimated carrying capacity %.0f)",
			requestedInitialPopulation,
			seededPopulation,
			metrics.CarryingCapacity,
		)
	}

	result.Samples = append(
		result.Samples,
		buildPopulationSample(0, terrainMap, populationModel, intervalAccumulator{}),
	)

	accumulator := intervalAccumulator{}
	for year := 1; year <= e.config.Years; year++ {
		tick := populationModel.AdvanceOneYear(terrainMap)
		accumulator.addTick(tick)

		if shouldSample(year, e.config.Years, e.config.Interval) {
			result.Samples = append(
				result.Samples,
				buildPopulationSample(year, terrainMap, populationModel, accumulator),
			)
			accumulator = intervalAccumulator{}
		}
	}

	return result, nil
}

func generateExperimentWorld(seed string) (*simtypes.TerrainMap, error) {
	seedHash := sha256.Sum256([]byte(seed))
	random := rng.NewRNG(seedHash)
	pipeline := layers.NewPipeline(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight, random)
	terrainMap, err := pipeline.GenerateWorld(seedHash)
	if err != nil {
		return nil, err
	}
	terrainMap.SetInitialized(true)
	return terrainMap, nil
}

func shouldSample(year, totalYears, interval int) bool {
	return year == totalYears || year%interval == 0
}

func buildPopulationSample(
	year int,
	terrainMap *simtypes.TerrainMap,
	populationModel *PopulationModel,
	acc intervalAccumulator,
) PopulationSample {
	metrics := CollectPopulationMetrics(terrainMap, populationModel)
	averageDistance := 0.0
	if acc.migratedPopulation > 0 {
		averageDistance = acc.distanceWeight / acc.migratedPopulation
	}

	return PopulationSample{
		Year:               year,
		Population:         metrics.Population,
		CarryingCapacity:   metrics.CarryingCapacity,
		Utilization:        metrics.Utilization,
		OccupiedCells:      metrics.OccupiedCells,
		OverCapacityCells:  metrics.OverCapacityCells,
		ByTerrain:          toTerrainPopulationSamples(metrics.ByTerrain),
		MigratedPopulation: acc.migratedPopulation,
		SourceCells:        acc.sourceCells,
		DestinationCells:   acc.destinationCells,
		AverageDistance:    averageDistance,
		MaxDistance:        acc.maxDistance,
		StarvationDeaths:   acc.starvationDeaths,
	}
}

func (acc *intervalAccumulator) addTick(tick PopulationTickResult) {
	acc.migratedPopulation += tick.MigratedPopulation
	acc.sourceCells += tick.SourceCells
	acc.destinationCells += tick.DestinationCells
	acc.distanceWeight += tick.AverageDistance * tick.MigratedPopulation
	if tick.MaxDistance > acc.maxDistance {
		acc.maxDistance = tick.MaxDistance
	}
	acc.starvationDeaths += tick.StarvationDeaths
}

func toTerrainPopulationSamples(metrics []TerrainPopulationMetrics) []TerrainPopulationSample {
	samples := make([]TerrainPopulationSample, 0, len(metrics))
	for _, metric := range metrics {
		samples = append(samples, TerrainPopulationSample{
			Terrain:          metric.Terrain,
			Population:       metric.Population,
			CarryingCapacity: metric.CarryingCapacity,
			Utilization:      metric.Utilization,
			OccupiedCapacity: metric.OccupiedCapacity,
			OccupiedRatio:    metric.OccupiedRatio,
		})
	}
	return samples
}
