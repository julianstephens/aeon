package layers

import "github.com/julianstephens/aeon/internal/simtypes"

type WorldDiagnostics struct {
	TerrainCounts          map[simtypes.TerrainType]int
	TerrainPercentages     map[simtypes.TerrainType]float64
	PassableLandPercentage float64

	LargestRegion             map[simtypes.TerrainType]int
	RegionCounts              map[simtypes.TerrainType]int
	LargestPassableLandRegion int
}

func ComputeWorldDiagnostics(tm simtypes.TerrainMap) WorldDiagnostics {
	totalCells := tm.Width * tm.Height
	terrainCounts := map[simtypes.TerrainType]int{}
	terrainPercentages := map[simtypes.TerrainType]float64{}
	largestRegion := map[simtypes.TerrainType]int{}
	regionCounts := map[simtypes.TerrainType]int{}

	for _, terrainType := range terrainTypes() {
		terrainCounts[terrainType] = 0
		terrainPercentages[terrainType] = 0
		largestRegion[terrainType] = 0
		regionCounts[terrainType] = 0
	}

	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			cell := tm.GetCell(x, y)
			if cell == nil {
				continue
			}
			terrainCounts[cell.Terrain]++
		}
	}

	if totalCells > 0 {
		for _, terrainType := range terrainTypes() {
			terrainPercentages[terrainType] = float64(terrainCounts[terrainType]) / float64(totalCells)
		}
	}

	for _, terrainType := range terrainTypes() {
		sizes := connectedRegionSizes(tm, func(cell *simtypes.TerrainCell) bool {
			return cell.Terrain == terrainType
		})

		regionCounts[terrainType] = len(sizes)
		for _, size := range sizes {
			if size > largestRegion[terrainType] {
				largestRegion[terrainType] = size
			}
		}
	}

	passableCount := terrainCounts[simtypes.TerrainTypePlains] + terrainCounts[simtypes.TerrainTypeForest]
	passableLandPercentage := 0.0
	if totalCells > 0 {
		passableLandPercentage = float64(passableCount) / float64(totalCells)
	}

	largestPassableLandRegion := 0
	passableRegionSizes := connectedRegionSizes(tm, func(cell *simtypes.TerrainCell) bool {
		return cell.Terrain.IsPassable()
	})
	for _, size := range passableRegionSizes {
		if size > largestPassableLandRegion {
			largestPassableLandRegion = size
		}
	}

	return WorldDiagnostics{
		TerrainCounts:             terrainCounts,
		TerrainPercentages:        terrainPercentages,
		PassableLandPercentage:    passableLandPercentage,
		LargestRegion:             largestRegion,
		RegionCounts:              regionCounts,
		LargestPassableLandRegion: largestPassableLandRegion,
	}
}

func connectedRegionSizes(tm simtypes.TerrainMap, include func(cell *simtypes.TerrainCell) bool) []int {
	visited := make([]bool, tm.Width*tm.Height)
	sizes := make([]int, 0)

	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			start := y*tm.Width + x
			if visited[start] {
				continue
			}

			cell := tm.GetCell(x, y)
			if cell == nil || !include(cell) {
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
					if neighbor.X < 0 || neighbor.X >= tm.Width || neighbor.Y < 0 || neighbor.Y >= tm.Height {
						continue
					}

					index := neighbor.Y*tm.Width + neighbor.X
					if visited[index] {
						continue
					}

					neighborCell := tm.GetCell(neighbor.X, neighbor.Y)
					if neighborCell == nil || !include(neighborCell) {
						visited[index] = true
						continue
					}

					visited[index] = true
					queue = append(queue, neighbor)
				}
			}

			sizes = append(sizes, size)
		}
	}

	return sizes
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
