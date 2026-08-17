# Aeon

Aeon is a deterministic simulation of a small society evolving over time. It is not intended to be a realistic civilization simulator. Its purpose is to create a small world whose behavior becomes difficult to predict once its rules begin interacting.

## Model

An Aeon world consists of:

* A fixed rectangular map
* Terrain and local resources
* Individual agents
* Settlements
* Environmental conditions
* Historical events
* A simulation clock

Agents have a small set of properties:

* Age
* Health
* Food
* Wealth
* Location
* Occupation
* Reproductive status
* Individual biological modifiers

The simulation advances one year at a time.

A year's simulation follows an explicit pipeline:

1. Environment changes
2. Resources are produced
3. Agents consume resources
4. Agents move
5. Agents reproduce
6. Mortality is resolved
7. Settlements are updated
8. Historical events are recorded
9. The year advances

The rules are intentionally simple. Complexity should emerge from their interaction rather than from sophisticated individual agents.

## Determinism

Every world has a seed.

All randomness used by the simulation originates from the world's deterministic random number generator. Consequently:

```text
same seed
    +
same initial configuration
    +
same sequence of simulation steps
    =
same world
```

This makes worlds reproducible and allows interesting simulations to be shared by seed.

For example:

```bash
aeon simulate --seed 42 --years 1000
```

should produce the same result every time it is run.

## Architecture

The simulation is deliberately separated from transport and presentation.

```text
┌──────────────────────┐
│       Frontend       │
└──────────┬───────────┘
           │ HTTP
┌──────────▼───────────┐
│       API layer      │
└──────────┬───────────┘
           │
┌──────────▼───────────┐
│  Simulation engine   │
│                      │
│  World               │
│  Geography           │
│  Agents              │
│  Settlements         │
│  Environment         │
│  Events              │
│  RNG                 │
└──────────────────────┘
```

The simulation engine should not depend on the HTTP server, frontend, or persistence layer.

## Development

Requirements:

* Go

Run the tests:

```bash
go test ./...
```

### Experiment runner

Run a repeatable population experiment in text mode:

```bash
go run ./cmd/aeon experiment --scenario crowded --seed 42 --years 100 --interval 25
```

Expected output shape:

```text
Aeon Population Experiment

Seed:       42
Years:      100
Population: 10000
Scenario:   crowded

Year  Population  Capacity  Util.  Cells  Migrants  Deaths
0     10000       27954     35.8%  959    0         0
25    10781       27954     38.6%  1788   3180      0
50    12168       27954     43.5%  1793   2702      0
75    13732       27954     49.1%  1835   2782      0
100   15177       27954     54.3%  1978   3285      0
```

Run the same experiment in JSON mode:

```bash
go run ./cmd/aeon experiment --scenario crowded --seed 42 --years 100 --interval 25 --format json
```

Expected JSON shape:

```json
{
  "config": {
    "Seed": "42",
    "Years": 100,
    "InitialPopulation": 10000,
    "GrowthRate": 0.025,
    "StarvationRate": 0.1,
    "MigrationRate": 0.07,
    "MaxCellCapacity": 30,
    "Interval": 25
  },
  "samples": [
    {
      "year": 0,
      "population": 10000,
      "carrying_capacity": 27954.470598720287,
      "utilization": 0.3577245351395703,
      "occupied_cells": 959,
      "over_capacity_cells": 0,
      "migrated_population": 0,
      "source_cells": 0,
      "destination_cells": 0,
      "average_distance": 0,
      "max_distance": 0,
      "starvation_deaths": 0
    }
  ]
}
```

Notes:

* Scenarios: `baseline`, `crowded`, `isolated`, `scarcity`
* CLI flags override scenario values (for example, `--population` or `--migration-rate`)
* For `years=100` and `interval=25`, samples are recorded at years `0, 25, 50, 75, 100`
* Results are deterministic for the same seed and configuration

Run a simulation:

```bash
aeon simulate --seed 42 --years 100
```

Start the API:

```bash
aeon serve
```

## API

The initial API is intentionally small.

### Health

```http
GET /health
```

### Create a world

```http
POST /api/worlds
Content-Type: application/json

{
  "seed": 42
}
```

### Get a world

```http
GET /api/worlds/{id}
```

### Advance a world

```http
POST /api/worlds/{id}/step
Content-Type: application/json

{
  "years": 10
}
```

Worlds are currently held in memory. Restarting the process destroys them.
