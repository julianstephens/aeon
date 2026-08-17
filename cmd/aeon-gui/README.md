# Aeon GUI v1

This is the Ebitengine inspector for the Aeon simulation engine. It is intentionally a viewer/controller over the deterministic simulation layer, not a second simulation implementation.

## Run

```bash
go run ./cmd/aeon-gui --seed 42 --population 1000
```

## Controls

- Space: play/pause
- Right arrow: advance one simulated year
- R: reset to the original configuration
- 1–6: switch map layer
  - 1 Terrain
  - 2 Elevation
  - 3 Moisture
  - 4 Fertility
  - 5 Food
  - 6 Population
- Left click: select a cell in the world map

## Notes

- The app starts directly in inspection mode with a fixed logical viewport of 1280x800.
- Simulation time is driven by the in-memory year state, not real time.
- Reset rebuilds a fresh simulation from the same seed and parameters so the initial state is identical to a fresh launch.

## Headless / smoke testing

This repo keeps the simulation logic in `internal/simulation` and renders through Ebitengine. For automated testing, the GUI can be driven headlessly with a deterministic tick-based harness to:

- launch the app
- inject keyboard and mouse input
- advance one or many ticks
- capture rendered frames
- compare against a baseline image or a lightweight state assertion

The current codebase lays the simulation-side foundation for that approach without introducing a large visual test suite before the UI stabilizes.
