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
	seed             [32]byte
	random           *rng.RNG
	terrainGenerator *layers.Generator

	CurrentYear      int
	terrainMap       simtypes.TerrainMap
	agents           []*Agent
	settlements      []Settlement
	historicalEvents []string
}

func NewWorld(seed string) (*World, error) {
	logger.WithFields(map[string]interface{}{
		"seed":   seed,
		"width":  simtypes.DefaultMapWidth,
		"height": simtypes.DefaultMapHeight,
	}).Debug("creating new world")

	terrainMap := simtypes.NewTerrainMap(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight)
	seedHash := sha256.Sum256([]byte(seed))
	random := rng.NewRNG(seedHash)
	world := &World{
		seed:             seedHash,
		random:           random,
		CurrentYear:      0,
		terrainMap:       terrainMap,
		terrainGenerator: layers.NewGenerator(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight, random),
		agents:           []*Agent{},
		settlements:      []Settlement{},
		historicalEvents: []string{},
	}
	if err := world.initialize(); err != nil {
		logger.Errorf("world initialization failed: %v", err)
		return nil, err
	}

	logger.WithFields(map[string]interface{}{
		"current_year": world.CurrentYear,
		"population":   world.PopulationCount(),
	}).Debug("world created")
	return world, nil
}

func (w *World) initialize() error {
	logger.Debug("initializing world terrain map")
	if err := generateTerrainMap(w); err != nil {
		return &SimulationError{Code: CodeWorldError, Message: "Failed to generate terrain map", Cause: err}
	}
	logger.Debug("terrain map initialized")

	logger.WithField("initial_agents", 100).Debug("generating initial population")
	if err := generateInitialPopulation(w, 100); err != nil {
		return &SimulationError{Code: CodeWorldError, Message: "Failed to generate initial population", Cause: err}
	}
	logger.WithField("population", w.PopulationCount()).Debug("initial population generated")
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
	return len(w.agents)
}

func (w *World) PrintSummary(c *cliutil.Console) {
	_ = c.Info(fmt.Sprintf("World: %x", w.seed))
	_ = c.PrintColored(fmt.Sprintf("Year: %d", w.CurrentYear), cliutil.ColorWhite)
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

func generateTerrainMap(w *World) (err error) {
	err = w.terrainGenerator.GenerateTerrainMap(w.seed, &w.terrainMap)
	return
}

func generateInitialPopulation(w *World, numAgents int) (err error) {
	for i := range numAgents {
		agent, agentErr := NewAgent(
			w.random.AgeInRange(0, 70),
			w.random.Productivity(),
			w.random.HealthInRange(50, 100),
			w.random.HealthResilience(),
			w.random.FoodCapacityInRange(10, 80),
			w.random.WealthInRange(0, 200),
			w.random.Sex(),
			w.random.Fertility(),
			w.random.Location(),
			w.random.Occupation(),
		)
		if agentErr != nil {
			err = agentErr
			return
		}
		w.AddAgent(agent)

		if (i+1)%25 == 0 || i+1 == numAgents {
			logger.WithFields(map[string]interface{}{
				"generated": i + 1,
				"target":    numAgents,
			}).Debug("initial population generation progress")
		}
	}
	return
}
