package simulation

import (
	"github.com/julianstephens/go-utils/cliutil"
	"github.com/julianstephens/go-utils/logger"
)

// Run runs the simulation for a given number of years and prints the results to the console.
func Run(c *cliutil.Console, seed string, years int) error {
	logger.WithFields(map[string]interface{}{
		"seed":  seed,
		"years": years,
	}).Debug("starting simulation run")

	// Initialize the world with the given seed
	w, err := NewWorld(seed)
	if err != nil {
		logger.Errorf("failed to initialize world: %v", err)
		return err
	}

	logger.WithFields(map[string]interface{}{
		"year":       w.currentYear,
		"population": w.PopulationCount(),
	}).Debug("world initialized")
	w.PrintSummary(c)
	logger.Debug("simulation run completed")
	return nil
}
