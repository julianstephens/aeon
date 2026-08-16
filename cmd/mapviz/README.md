# mapviz

`mapviz` is a small development tool for inspecting Aeon's terrain layers as PNG images.

It is intended for debugging terrain generation rather than as part of the runtime simulation or web application.

## Usage

Run from the repository root:

```bash
go run ./cmd/mapviz
```

By default this generates `terrain.png` using seed `42`, renders the elevation layer, and scales each map cell by 8 pixels.

### Options

```text
-seed uint
    Terrain generation seed (default 42)

-layer string
    Layer to render (default "elevation")

-scale int
    Scale factor for each map cell (default 8)

-output string
    Output PNG path (default "terrain.png")
```

For example:

```bash
go run ./cmd/mapviz \
  -seed 42 \
  -layer elevation \
  -scale 8 \
  -output elevation.png
```

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
go run ./cmd/mapviz -seed 42 -layer terrain -output terrain.png
go run ./cmd/mapviz -seed 42 -layer elevation -output elevation.png
```

The scalar layers are expected to contain normalized values in the range `[0, 1]`. Values outside that range are clamped for visualization.

## Reproducibility

Terrain generation is deterministic for a given seed. This makes `mapviz` useful for comparing changes to generation algorithms:

```bash
go run ./cmd/mapviz -seed 42 -layer elevation -output before.png
```

After changing the generator, run the same command again and compare the resulting image.

Changing the seed produces a different deterministic world:

```bash
go run ./cmd/mapviz -seed 1337 -layer elevation -output world-1337.png
```

## What to Look For

When inspecting the elevation layer, look for broad, spatially coherent features rather than isolated pixel-scale variation.

For diamond-square generation in particular, the elevation image should make it easy to spot:

- large-scale elevation gradients
- clustered high and low areas
- abrupt or unintended discontinuities
- boundary artifacts
- excessive uniformity
- excessive high-frequency noise

Inspect intermediate layers independently before relying on the final terrain classification. This helps distinguish problems in layer generation from problems in terrain classification.

## Scope

`mapviz` intentionally has no persistence, server, or frontend dependencies. It should remain a small command-line debugging tool for terrain-generation work.
