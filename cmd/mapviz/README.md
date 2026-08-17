# mapviz

`mapviz` is a small development tool for inspecting Aeon's terrain layers as PNG images.

It is intended for debugging terrain generation rather than as part of the runtime simulation or web application.

## Usage

Run from the repository root:

```bash
go run ./cmd/mapviz --help
```

Current CLI:

```text
Usage: mapviz <command>

Commands:
  analyze    Analyze generated terrain layers.
  render     Render a generated terrain layer to PNG.
```

## Analyze

Analyze prints scalar stats and terrain diagnostics for a deterministic generated world:

```bash
go run ./cmd/mapviz analyze --seed 42
```

Options:

```text
--seed=42
    Terrain generation seed.
```

The output includes:

- elevation, moisture, and fertility min/max/mean/stddev
- terrain distribution by type
- connected-region counts and largest region per terrain
- world viability summary and reasons when invalid
- terrain boundary diagnostics
- neighbor agreement diagnostics

`analyze` uses the same terrain-generation path as simulation (`Pipeline.GenerateWorld`) and the same viability rules (`DefaultViabilityRules`) from the layers package.

## Render

Render produces a PNG for one layer:

```bash
go run ./cmd/mapviz render --seed 42 --layer elevation --scale 8 --output elevation.png
```

Options:

```text
--seed=42
    Terrain generation seed.

--layer="elevation"
    Layer to render. One of: terrain, elevation, moisture, fertility, food-capacity.

--scale=8
    Scale factor for each map cell.

--output="terrain.png"
    Output PNG path.
```

Notes:

- output path must be relative
- output path must end in `.png`
- scale must be at least `1`

## Layers

The following layer types are currently supported:

| Layer | Description |
| --- | --- |
| `terrain` | Render the classified terrain map using discrete terrain colors. |
| `elevation` | Render elevation as a grayscale image, with low elevation dark and high elevation light. |
| `moisture` | Render cell moisture as a grayscale image. |
| `fertility` | Render cell fertility as a grayscale image. |
| `food-capacity` | Render cell food capacity as a grayscale image. |

For example:

```bash
go run ./cmd/mapviz render --seed 42 --layer terrain --output terrain.png
go run ./cmd/mapviz render --seed 42 --layer elevation --output elevation.png
```

The scalar layers are expected to contain normalized values in the range `[0, 1]`. Values outside that range are clamped for visualization.

## Reproducibility

Terrain generation is deterministic for a given seed. This makes `mapviz` useful for comparing changes to generation algorithms:

```bash
go run ./cmd/mapviz render --seed 42 --layer elevation --output before.png
```

After changing the generator, run the same command again and compare the resulting image.

Changing the seed produces a different deterministic world:

```bash
go run ./cmd/mapviz render --seed 1337 --layer elevation --output world-1337.png
```

## Viability Output

`mapviz analyze` includes a world viability section:

```text
World viability
  Passable land:        76.4%
  Largest land region:  68.2%
  Water:                23.6%
  Mountain:              6.8%
  Viable:               yes
```

If a generated attempt fails validation, analyze reports why:

```text
World viability
  Passable land:        12.4%
  Largest land region:   7.1%
  Water:                81.3%
  Mountain:              6.3%
  Viable:               no

Reasons:
  - passable land below 25.0%
  - water exceeds 60.0%
```


