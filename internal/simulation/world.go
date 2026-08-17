package simulation

import (
	"crypto/sha256"
	"fmt"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/layers"
	"github.com/julianstephens/aeon/internal/simulation/rng"
	"github.com/julianstephens/go-utils/cliutil"
	"github.com/julianstephens/go-utils/generic"
	"github.com/julianstephens/go-utils/logger"
)

type Settlement struct {
	Name     string
	Location simtypes.Position
	Agents   []Agent
}

type World struct {
	seed   [32]byte
	random *rng.RNG

	terrainMap    *simtypes.TerrainMap
	layerPipeline *layers.Pipeline

	populationModel *PopulationModel

	agents           []*Agent
	settlements      []Settlement
	historicalEvents []string
}

func NewWorld(seed string, params PopulationParameters) (*World, error) {
	logger.WithFields(map[string]any{
		"seed":   seed,
		"width":  simtypes.DefaultMapWidth,
		"height": simtypes.DefaultMapHeight,
	}).Debug("creating new world")

	seedHash := sha256.Sum256([]byte(seed))
	random := rng.NewRNG(seedHash)
	world := &World{
		seed:             seedHash,
		random:           random,
		layerPipeline:    layers.NewPipeline(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight, random),
		terrainMap:       nil,
		agents:           []*Agent{},
		settlements:      []Settlement{},
		historicalEvents: []string{},
		populationModel:  NewPopulationModel(params),
	}
	if err := world.initialize(); err != nil {
		logger.Errorf("world initialization failed: %v", err)
		return nil, err
	}

	logger.WithFields(map[string]any{
		"current_year": world.Year(),
		"population":   world.PopulationCount(),
	}).Debug("world created")
	return world, nil
}

func (w *World) initialize() error {
	logger.Debug("initializing world terrain map")
	tm, err := w.layerPipeline.GenerateWorld(w.seed)
	if err != nil || tm == nil {
		return &SimulationError{Code: CodeWorldError, Message: "Failed to generate terrain map", Cause: err}
	}
	w.terrainMap = tm
	logger.Debug("terrain map initialized")
	w.terrainMap.SetInitialized(true)

	logger.WithFields(map[string]any{
		"year":       w.Year(),
		"population": w.PopulationCount(),
	}).Debug("seeding initial population")
	seeded := w.populationModel.SeedPopulation(w.terrainMap, InitialPopulationCount)
	logger.WithFields(map[string]any{
		"year":       w.Year(),
		"population": w.PopulationCount(),
		"seeded":     seeded,
	}).Debug("population seeded")
	return nil
}

func (w *World) AddAgent(a *Agent) {
	w.agents = append(w.agents, a)
}

func (w *World) RemoveAgent(id string) {
	w.agents = generic.Filter(w.agents, func(a *Agent) bool { return a.ID != id })
}

func (w *World) AddSettlement(settlement Settlement) {
	w.settlements = append(w.settlements, settlement)
}

func (w *World) RemoveSettlement(name string) {
	w.settlements = generic.Filter(w.settlements, func(s Settlement) bool { return s.Name != name })
}

func (w *World) PopulationCount() int {
	return w.populationModel.TotalPopulation(w.terrainMap)
}

func (w *World) Year() int {
	return w.populationModel.CurrentTick()
}

func (w *World) PrintSummary(c *cliutil.Console) {
	_ = c.Info(fmt.Sprintf("World: %x", w.seed))
	_ = c.PrintColored(fmt.Sprintf("Year: %d", w.Year()), cliutil.ColorWhite)
	_ = c.PrintColored(fmt.Sprintf("Population: %d", w.PopulationCount()), cliutil.ColorWhite)
}

func (w *World) PrintPopulationDetails() {
	table := [][]string{
		{"ID", "Age", "Health", "Stored Food", "Wealth", "Sex", "Location", "Occupation", "Is Alive", "Can Reproduce"},
	}
	for _, agent := range w.agents {
		table = append(table, []string{
			agent.ID,
			fmt.Sprintf("%d", agent.Age),
			fmt.Sprintf("%.2f", agent.Health),
			fmt.Sprintf("%.2f", agent.Food),
			fmt.Sprintf("%.2f", agent.Wealth),
			fmt.Sprintf("%d", agent.Sex),
			fmt.Sprintf("(%d, %d)", agent.Location.X, agent.Location.Y),
			agent.Occupation.String(),
			fmt.Sprintf("%t", agent.IsAlive),
		})
	}
	cliutil.PrintTable(table)
}
