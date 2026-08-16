package simulation

import "fmt"

type SimulationError struct {
	Code    string
	Message string
	Cause   error
}

const (
	CodeWorldError = "WORLD_ERROR"
	CodeAgentError = "AGENT_ERROR"
)

func (s *SimulationError) Error() string {
	if s.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", s.Code, s.Message, s.Cause)
	}
	return fmt.Sprintf("%s: %s", s.Code, s.Message)
}

func (s *SimulationError) Unwrap() error {
	return s.Cause
}
