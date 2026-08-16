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
