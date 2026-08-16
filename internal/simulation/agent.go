package simulation

import (
	"fmt"

	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/go-utils/cliutil"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Agent struct {
	ID               string
	Age              uint8
	Health           float64
	Food             float64
	Wealth           float64
	Fertility        float64
	Productivity     float64
	HealthResilience float64

	Sex         simtypes.Sex
	Occupation  simtypes.Occupation
	Location    simtypes.Position
	HouseholdID uint64
	IsAlive     bool
}

func NewAgent(
	age uint8,
	productivity float64,
	health float64,
	healthResilience float64,
	food float64,
	wealth float64,
	sex simtypes.Sex,
	fertility float64,
	location simtypes.Position,
	occupation simtypes.Occupation,
) (*Agent, error) {
	id, err := gonanoid.New()
	if err != nil {
		return nil, &SimulationError{
			Code:    CodeAgentError,
			Message: "failed to generate agent ID",
			Cause:   err,
		}
	}

	return &Agent{
		ID:               id,
		Age:              age,
		Health:           health,
		Food:             food,
		Wealth:           wealth,
		Fertility:        fertility,
		Productivity:     productivity,
		HealthResilience: healthResilience,
		Sex:              sex,
		Location:         location,
		Occupation:       occupation,
		IsAlive:          true,
	}, nil
}

func (a *Agent) PrintInfo() {
	table := [][]string{
		{"ID", "Age", "Health", "Food", "Wealth", "Sex", "Location", "Occupation", "Is Alive"},
		{
			a.ID,
			fmt.Sprintf("%d", a.Age),
			fmt.Sprintf("%.2f", a.Health),
			fmt.Sprintf("%.2f", a.Food),
			fmt.Sprintf("%.2f", a.Wealth),
			fmt.Sprintf("%d", a.Sex),
			fmt.Sprintf("(%d, %d)", a.Location.X, a.Location.Y),
			a.Occupation.String(),
			fmt.Sprintf("%t", a.IsAlive),
		},
	}
	cliutil.PrintTable(table)
}
