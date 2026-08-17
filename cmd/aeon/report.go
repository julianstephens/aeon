package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"text/tabwriter"

	"github.com/julianstephens/aeon/internal/simulation"
)

func renderExperimentText(scenario simulation.Scenario, result simulation.ExperimentResult) string {
	var buffer bytes.Buffer
	writer := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

	fmt.Fprintln(&buffer, "Aeon Population Experiment")
	fmt.Fprintln(&buffer)
	fmt.Fprintf(&buffer, "Seed:\t%s\n", result.Config.Seed)
	fmt.Fprintf(&buffer, "Years:\t%d\n", result.Config.Years)
	fmt.Fprintf(&buffer, "Population:\t%s\n", formatRounded(result.Config.InitialPopulation))
	fmt.Fprintf(&buffer, "Scenario:\t%s\n", scenario)
	fmt.Fprintln(&buffer)

	_, _ = fmt.Fprintln(writer, "Year\tPopulation\tCapacity\tUtil.\tCells\tMigrants\tDeaths")
	for _, sample := range result.Samples {
		_, _ = fmt.Fprintf(
			writer,
			"%d\t%s\t%s\t%.1f%%\t%d\t%s\t%s\n",
			sample.Year,
			formatRounded(sample.Population),
			formatRounded(sample.CarryingCapacity),
			sample.Utilization*100,
			sample.OccupiedCells,
			formatRounded(sample.MigratedPopulation),
			formatRounded(sample.StarvationDeaths),
		)
	}
	_ = writer.Flush()

	if len(result.Samples) > 0 {
		final := result.Samples[len(result.Samples)-1]
		totalDeaths := 0.0
		totalMigrated := 0.0
		for _, sample := range result.Samples {
			totalDeaths += sample.StarvationDeaths
			totalMigrated += sample.MigratedPopulation
		}

		fmt.Fprintln(&buffer)
		fmt.Fprintln(&buffer, "Final")
		fmt.Fprintf(&buffer, "  population:\t%s\n", formatRounded(final.Population))
		fmt.Fprintf(&buffer, "  utilization:\t%.1f%%\n", final.Utilization*100)
		fmt.Fprintf(&buffer, "  occupied cells:\t%d\n", final.OccupiedCells)
		fmt.Fprintf(&buffer, "  starvation deaths:\t%s\n", formatRounded(totalDeaths))
		fmt.Fprintf(&buffer, "  migrated population:\t%s\n", formatRounded(totalMigrated))
	}

	return buffer.String()
}

func renderExperimentJSON(result simulation.ExperimentResult) (string, error) {
	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to encode experiment result as JSON: %w", err)
	}
	return string(payload), nil
}

func formatRounded(value float64) string {
	rounded := int64(math.Round(value))
	return strconv.FormatInt(rounded, 10)
}
