package simulation

import "github.com/julianstephens/go-utils/cliutil"

// Run runs the simulation for a given number of years and prints the results to the console.
func Run(c *cliutil.Console, seed string, years int) error {
	// Initialize the world with the given seed
	w, err := NewWorld(seed)
	if err != nil {
		return err
	}
	w.PrintSummary(c)
	return nil
}
