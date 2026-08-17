// Package simulation contains Aeon's deterministic world simulation.
//
// The package models the state of a small world and advances that world through
// discrete yearly simulation steps. It owns the simulation rules, world state,
// deterministic random number generation, geography, agents, settlements, and
// historical events.
//
// Simulation state is independent of transport and presentation concerns. Code
// outside this package should interact with a World through its public API
// rather than depending on internal simulation details.
//
// Given the same initial configuration and seed, a simulation must produce the
// same state after the same sequence of steps.
package simulation
