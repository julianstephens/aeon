package main

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/julianstephens/aeon/internal/simtypes"
	"github.com/julianstephens/aeon/internal/simulation/layers"
	"github.com/julianstephens/aeon/internal/simulation/rng"
)

type LayerType string

const (
	LayerTypeTerrain      LayerType = "terrain"
	LayerTypeElevation    LayerType = "elevation"
	LayerTypeMoisture     LayerType = "moisture"
	LayerTypeFertility    LayerType = "fertility"
	LayerTypeFoodCapacity LayerType = "food-capacity"
)

type CLI struct {
	Analyze AnalyzeCommand `cmd:"" help:"Analyze generated terrain layers."`
	Render  RenderCommand  `cmd:"" help:"Render a generated terrain layer to PNG."`
}

type AnalyzeCommand struct {
	Seed uint64 `help:"Terrain generation seed." default:"42"`
}

type RenderCommand struct {
	Seed   uint64    `help:"Terrain generation seed."        default:"42"`
	Layer  LayerType `help:"Layer to render."                default:"elevation"   enum:"terrain,elevation,moisture,fertility,food-capacity"`
	Scale  int       `help:"Scale factor for each map cell." default:"8"`
	Output string    `help:"Output PNG path."                default:"terrain.png"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("mapviz"),
		kong.Description("Generate and analyze Aeon terrain layers."),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
	)

	switch ctx.Command() {
	case "analyze":
		if err := analyze(cli.Analyze.Seed); err != nil {
			ctx.Fatalf("analysis failed: %v", err)
		}
	case "render":
		if err := render(cli.Render); err != nil {
			ctx.Fatalf("render failed: %v", err)
		}
	}
}

func render(cli RenderCommand) error {
	if cli.Scale < 1 {
		return fmt.Errorf("scale must be at least 1")
	}

	outputPath, err := sanitizeOutputPath(cli.Output)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	tm, elevation, err := generateTerrainMap(cli.Seed)
	if err != nil {
		return fmt.Errorf("failed to generate terrain map: %w", err)
	}

	img, err := renderLayer(*tm, elevation, cli.Layer, cli.Scale)
	if err != nil {
		return fmt.Errorf("failed to render layer: %w", err)
	}

	if outputDir := filepath.Dir(outputPath); outputDir != "." {
		if err := os.MkdirAll(outputDir, 0o750); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	file, err := os.Create(outputPath) // #nosec G304 -- outputPath is restricted to relative .png paths.
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	fmt.Printf("wrote %s\n", outputPath)
	return nil
}

func analyze(seed uint64) error {
	tm, elevation, err := generateTerrainMap(seed)
	if err != nil {
		return err
	}

	fmt.Printf("Aeon Terrain Analysis\n")
	fmt.Printf("Seed: %d\n", seed)
	fmt.Printf("Map: %dx%d\n\n", tm.Width, tm.Height)

	printScalarStats("Elevation", *elevation)
	printScalarStats(
		"Moisture",
		*layerFromTerrain(*tm, func(cell *simtypes.TerrainCell) float64 { return cell.Moisture }),
	)
	printScalarStats(
		"Fertility",
		*layerFromTerrain(*tm, func(cell *simtypes.TerrainCell) float64 { return cell.Fertility }),
	)
	printTerrainDistribution(*tm)
	printConnectedRegions(*tm)
	printTerrainBoundaries(*tm)
	printNeighborAgreement(*tm)

	return nil
}

func printScalarStats(name string, layer simtypes.Layer) {
	minValue := math.Inf(1)
	maxValue := math.Inf(-1)
	var sum float64
	count := layer.Width * layer.Height

	for y := 0; y < layer.Height; y++ {
		for x := 0; x < layer.Width; x++ {
			value := layer.Get(x, y)
			minValue = minFloat(minValue, value)
			maxValue = maxFloat(maxValue, value)
			sum += value
		}
	}

	mean := 0.0
	if count > 0 {
		mean = sum / float64(count)
	}

	var variance float64
	if count > 0 {
		for y := 0; y < layer.Height; y++ {
			for x := 0; x < layer.Width; x++ {
				delta := layer.Get(x, y) - mean
				variance += delta * delta
			}
		}
		variance /= float64(count)
	}

	fmt.Printf("%s\n", name)
	fmt.Printf("  min:    %.3f\n", minValue)
	fmt.Printf("  max:    %.3f\n", maxValue)
	fmt.Printf("  mean:   %.3f\n", mean)
	fmt.Printf("  stddev: %.3f\n\n", math.Sqrt(variance))
}

func printTerrainDistribution(tm simtypes.TerrainMap) {
	counts := terrainCounts(tm)
	total := tm.Width * tm.Height

	fmt.Printf("Terrain distribution\n")
	for _, terrain := range terrainTypes() {
		percent := 0.0
		if total > 0 {
			percent = float64(counts[terrain]) / float64(total) * 100
		}
		fmt.Printf("  %-9s %5.1f%% (%d)\n", terrain.String()+":", percent, counts[terrain])
	}
	fmt.Println()
}

func printConnectedRegions(tm simtypes.TerrainMap) {
	fmt.Printf("Connected regions\n")
	for _, terrain := range terrainTypes() {
		regions := connectedRegionSizes(tm, terrain)
		largest, isolated := 0, 0
		for _, size := range regions {
			largest = maxInt(largest, size)
			if size == 1 {
				isolated++
			}
		}
		fmt.Printf("  %-9s regions=%d largest=%d isolated=%d\n", terrain.String()+":", len(regions), largest, isolated)
	}
	fmt.Println()
}

func printTerrainBoundaries(tm simtypes.TerrainMap) {
	fmt.Printf("Terrain boundaries\n")
	total := tm.Width * tm.Height

	for _, terrain := range []simtypes.TerrainType{simtypes.TerrainTypeWater, simtypes.TerrainTypeMountain} {
		boundaryMask := makeBoundaryMask(tm, terrain)
		boundaryCells := countMask(boundaryMask)
		components := connectedMaskRegionSizes(boundaryMask, tm.Width, tm.Height)
		largest := 0
		for _, size := range components {
			largest = maxInt(largest, size)
		}

		percent := 0.0
		if total > 0 {
			percent = float64(boundaryCells) / float64(total) * 100
		}

		fmt.Printf(
			"  %-9s cells=%d (%5.1f%%) components=%d largest=%d\n",
			terrain.String()+":",
			boundaryCells,
			percent,
			len(components),
			largest,
		)
	}
	fmt.Println()
}

func makeBoundaryMask(tm simtypes.TerrainMap, target simtypes.TerrainType) []bool {
	mask := make([]bool, tm.Width*tm.Height)

	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			cell := tm.GetCell(x, y)
			if cell == nil || cell.Terrain != target {
				continue
			}

			for _, neighbor := range orthogonalNeighbors(simtypes.Position{X: x, Y: y}) {
				neighborCell := tm.GetCell(neighbor.X, neighbor.Y)
				if neighborCell != nil && neighborCell.Terrain != target {
					mask[y*tm.Width+x] = true
					break
				}
			}
		}
	}

	return mask
}

func countMask(mask []bool) int {
	count := 0
	for _, value := range mask {
		if value {
			count++
		}
	}
	return count
}

func connectedMaskRegionSizes(mask []bool, width, height int) []int {
	visited := make([]bool, len(mask))
	regions := make([]int, 0)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			start := y*width + x
			if visited[start] || !mask[start] {
				visited[start] = true
				continue
			}

			queue := []simtypes.Position{{X: x, Y: y}}
			visited[start] = true
			size := 0

			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]
				size++

				for _, neighbor := range orthogonalNeighbors(current) {
					if neighbor.X < 0 || neighbor.X >= width || neighbor.Y < 0 || neighbor.Y >= height {
						continue
					}

					index := neighbor.Y*width + neighbor.X
					if visited[index] || !mask[index] {
						continue
					}

					visited[index] = true
					queue = append(queue, neighbor)
				}
			}

			regions = append(regions, size)
		}
	}

	return regions
}

func connectedRegionSizes(tm simtypes.TerrainMap, target simtypes.TerrainType) []int {
	visited := make([]bool, tm.Width*tm.Height)
	regions := make([]int, 0)

	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			index := y*tm.Width + x
			if visited[index] {
				continue
			}

			cell := tm.GetCell(x, y)
			if cell == nil || cell.Terrain != target {
				visited[index] = true
				continue
			}

			queue := []simtypes.Position{{X: x, Y: y}}
			visited[index] = true
			size := 0

			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]
				size++

				for _, neighbor := range orthogonalNeighbors(current) {
					if neighbor.X < 0 || neighbor.X >= tm.Width || neighbor.Y < 0 || neighbor.Y >= tm.Height {
						continue
					}

					neighborIndex := neighbor.Y*tm.Width + neighbor.X
					if visited[neighborIndex] {
						continue
					}

					neighborCell := tm.GetCell(neighbor.X, neighbor.Y)
					if neighborCell == nil || neighborCell.Terrain != target {
						visited[neighborIndex] = true
						continue
					}

					visited[neighborIndex] = true
					queue = append(queue, neighbor)
				}
			}

			regions = append(regions, size)
		}
	}

	return regions
}

func printNeighborAgreement(tm simtypes.TerrainMap) {
	type counts struct {
		same  int
		total int
	}

	agreement := map[simtypes.TerrainType]counts{}

	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			cell := tm.GetCell(x, y)
			if cell == nil {
				continue
			}

			for _, neighbor := range orthogonalNeighbors(simtypes.Position{X: x, Y: y}) {
				neighborCell := tm.GetCell(neighbor.X, neighbor.Y)
				if neighborCell == nil {
					continue
				}

				stats := agreement[cell.Terrain]
				stats.total++
				if neighborCell.Terrain == cell.Terrain {
					stats.same++
				}
				agreement[cell.Terrain] = stats
			}
		}
	}

	fmt.Printf("Neighbor agreement\n")
	for _, terrain := range terrainTypes() {
		stats := agreement[terrain]
		percent := 0.0
		if stats.total > 0 {
			percent = float64(stats.same) / float64(stats.total) * 100
		}
		fmt.Printf("  %-9s %.1f%%\n", terrain.String()+":", percent)
	}
}

func terrainCounts(tm simtypes.TerrainMap) map[simtypes.TerrainType]int {
	counts := map[simtypes.TerrainType]int{}
	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			if cell := tm.GetCell(x, y); cell != nil {
				counts[cell.Terrain]++
			}
		}
	}
	return counts
}

func terrainTypes() []simtypes.TerrainType {
	return []simtypes.TerrainType{
		simtypes.TerrainTypeWater,
		simtypes.TerrainTypePlains,
		simtypes.TerrainTypeForest,
		simtypes.TerrainTypeMountain,
	}
}

func orthogonalNeighbors(position simtypes.Position) []simtypes.Position {
	return []simtypes.Position{
		{X: position.X - 1, Y: position.Y},
		{X: position.X + 1, Y: position.Y},
		{X: position.X, Y: position.Y - 1},
		{X: position.X, Y: position.Y + 1},
	}
}

func layerFromTerrain(tm simtypes.TerrainMap, valueAt func(*simtypes.TerrainCell) float64) *simtypes.Layer {
	layer := simtypes.NewLayer(tm.Width, tm.Height)
	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			if cell := tm.GetCell(x, y); cell != nil {
				layer.Set(x, y, valueAt(cell))
			}
		}
	}
	return layer
}

func generateTerrainMap(seed uint64) (*simtypes.TerrainMap, *simtypes.Layer, error) {
	seedBytes := sha256.Sum256(fmt.Appendf(nil, "%d", seed))
	random := rng.NewRNG(seedBytes)

	pipeline := layers.NewPipeline(simtypes.DefaultMapWidth, simtypes.DefaultMapHeight, random)
	tm, err := pipeline.Run(seedBytes)
	if err != nil {
		return nil, nil, err
	}

	return tm, layerFromTerrain(*tm, func(cell *simtypes.TerrainCell) float64 {
		return cell.Elevation
	}), nil
}

func renderLayer(tm simtypes.TerrainMap, elevation *simtypes.Layer, layer LayerType, scale int) (image.Image, error) {
	switch layer {
	case LayerTypeTerrain:
		return renderTerrainMap(tm, scale), nil
	case LayerTypeElevation:
		return renderScalarLayer(tm.Width, tm.Height, scale, func(x, y int) float64 {
			return elevation.Get(x, y)
		}), nil
	case LayerTypeMoisture:
		return renderScalarLayer(tm.Width, tm.Height, scale, func(x, y int) float64 {
			return tm.GetCell(x, y).Moisture
		}), nil
	case LayerTypeFertility:
		return renderScalarLayer(tm.Width, tm.Height, scale, func(x, y int) float64 {
			return tm.GetCell(x, y).Fertility
		}), nil
	case LayerTypeFoodCapacity:
		return renderScalarLayer(tm.Width, tm.Height, scale, func(x, y int) float64 {
			return tm.GetCell(x, y).FoodCapacity
		}), nil
	default:
		return nil, fmt.Errorf("unsupported layer %q", layer)
	}
}

func renderTerrainMap(tm simtypes.TerrainMap, scale int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, tm.Width*scale, tm.Height*scale))

	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			cell := tm.GetCell(x, y)
			if cell == nil {
				continue
			}
			fillCell(img, x, y, scale, terrainColor(cell.Terrain))
		}
	}

	return img
}

func renderScalarLayer(width, height, scale int, valueAt func(x, y int) float64) image.Image {
	img := image.NewGray(image.Rect(0, 0, width*scale, height*scale))
	minValue, maxValue := math.Inf(1), math.Inf(-1)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := valueAt(x, y)
			minValue = minFloat(minValue, value)
			maxValue = maxFloat(maxValue, value)
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := normalizeToRange(valueAt(x, y), minValue, maxValue)
			fillCell(img, x, y, scale, color.Gray{Y: uint8(value * 255)})
		}
	}

	return img
}

func fillCell(img interface{ Set(x, y int, c color.Color) }, x, y, scale int, c color.Color) {
	for dy := 0; dy < scale; dy++ {
		for dx := 0; dx < scale; dx++ {
			img.Set(x*scale+dx, y*scale+dy, c)
		}
	}
}

func terrainColor(terrain simtypes.TerrainType) color.Color {
	switch terrain {
	case simtypes.TerrainTypePlains:
		return color.RGBA{R: 190, G: 170, B: 100, A: 255}
	case simtypes.TerrainTypeForest:
		return color.RGBA{R: 50, G: 120, B: 60, A: 255}
	case simtypes.TerrainTypeMountain:
		return color.RGBA{R: 100, G: 100, B: 100, A: 255}
	case simtypes.TerrainTypeWater:
		return color.RGBA{R: 50, G: 100, B: 180, A: 255}
	default:
		return color.Black
	}
}

func normalizeToRange(value, minValue, maxValue float64) float64 {
	if maxValue <= minValue {
		return 0.5
	}

	normalized := (value - minValue) / (maxValue - minValue)
	if normalized < 0 {
		return 0
	}
	if normalized > 1 {
		return 1
	}
	return normalized
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sanitizeOutputPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}
	if filepath.Ext(path) != ".png" {
		return "", fmt.Errorf("output file must have .png extension")
	}
	cleanPath := filepath.Clean(path)
	if cleanPath == "." || cleanPath == ".." || cleanPath == "" {
		return "", fmt.Errorf("invalid output path")
	}
	return cleanPath, nil
}
