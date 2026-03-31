package main

import (
	"log"
	"math"
	"path/filepath"

	"github.com/wenta/timeseries-go/decompose"
	"github.com/wenta/timeseries-go/internal/exampleutil"
	"github.com/wenta/timeseries-go/plot"
)

func main() {
	outDir, err := exampleutil.OutputDir("decomposition")
	if err != nil {
		log.Fatalf("output dir creation failed: %v", err)
	}

	air, err := exampleutil.LoadCSVSeries("examples/data/air_passengers.csv", "2006-01-02", "air-passengers")
	if err != nil {
		log.Fatalf("air passengers csv load failed: %v", err)
	}

	// AirPassengers is close to multiplicative seasonality, so log-transforming
	// makes additive decomposition much easier to read.
	logAir := air.MapValues(func(v float64) float64 { return math.Log(v) })

	classical, err := decompose.SeasonalDecompose(logAir, 12)
	if err != nil {
		log.Fatalf("classical decomposition failed: %v", err)
	}

	stlResult, err := decompose.STL(logAir, decompose.STLConfig{Period: 12, Robust: true})
	if err != nil {
		log.Fatalf("stl decomposition failed: %v", err)
	}

	mustSave(outDir, "classical_trend", "Classical Decomposition: Observed vs Trend", []plot.Series{
		{Label: "log observed", Data: logAir, Color: plot.Gold},
		{Label: "trend", Data: classical.Trend, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
	})
	mustSave(outDir, "stl_trend", "STL: Observed vs Trend", []plot.Series{
		{Label: "log observed", Data: logAir, Color: plot.Gold},
		{Label: "trend", Data: stlResult.Trend, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
	})
	mustSave(outDir, "stl_seasonal", "STL Seasonal Component", []plot.Series{
		{Label: "seasonal", Data: stlResult.Seasonal, Color: plot.LightSeaGreen, Style: plot.LinePoints},
	})
	mustSave(outDir, "stl_residual", "STL Residual Component", []plot.Series{
		{Label: "residual", Data: stlResult.Residual, Color: plot.Crimson, Style: plot.LinePoints},
	})

	exampleutil.PrintOutputDir(outDir)
}

func mustSave(outDir string, slug string, title string, series []plot.Series) {
	if err := exampleutil.SaveAllFormats(
		filepath.Join(outDir, slug),
		series,
		plot.Title(title),
		plot.TimeFormat("2006"),
	); err != nil {
		log.Fatalf("%s plot failed: %v", slug, err)
	}
}
