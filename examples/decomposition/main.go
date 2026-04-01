package main

import (
	"log"
	"math"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/decompose"
	"github.com/wenta/timeseries-go/generator"
	"github.com/wenta/timeseries-go/internal/exampleutil"
	"github.com/wenta/timeseries-go/plot"
)

func main() {
	outDir, err := exampleutil.OutputDir("decomposition")
	if err != nil {
		log.Fatalf("output dir creation failed: %v", err)
	}
	report := exampleutil.NewReport("decomposition", "Decomposition Examples")

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

	mustSave(report, outDir, "classical_trend", "Classical Decomposition: Observed vs Trend", []plot.Series{
		{Label: "log observed", Data: logAir, Color: plot.Gold},
		{Label: "trend", Data: classical.Trend, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
	})
	mustSave(report, outDir, "stl_trend", "STL: Observed vs Trend", []plot.Series{
		{Label: "log observed", Data: logAir, Color: plot.Gold},
		{Label: "trend", Data: stlResult.Trend, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
	})
	mustSave(report, outDir, "stl_seasonal", "STL Seasonal Component", []plot.Series{
		{Label: "seasonal", Data: stlResult.Seasonal, Color: plot.LightSeaGreen, Style: plot.LinePoints},
	})
	mustSave(report, outDir, "stl_residual", "STL Residual Component", []plot.Series{
		{Label: "residual", Data: stlResult.Residual, Color: plot.Crimson, Style: plot.LinePoints},
	})

	hourlyBase := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	hourlyLength := 24 * 21
	hourlyIndex := generator.MakeSeriesIndex(hourlyBase, time.Hour, hourlyLength)
	hourlyValues := make([]float64, hourlyLength)
	for i := range hourlyValues {
		daily := 5 * math.Sin(2*math.Pi*float64(i%24)/24)
		weekly := []float64{-4, -2, 0, 1, 3, 5, -3}[(i/24)%7]
		trend := 50.0 + 0.05*float64(i)
		hourlyValues[i] = trend + daily + weekly
	}

	hourlySeries, err := timeseriesgo.Zip(hourlyIndex, hourlyValues)
	if err != nil {
		log.Fatalf("hourly multiseasonal series creation failed: %v", err)
	}

	mstlResult, err := decompose.MSTL(hourlySeries, decompose.MSTLConfig{
		Periods:         []int{24, 24 * 7},
		SeasonalWindows: []int{9, 15},
		Iterations:      2,
		InnerIterations: 2,
	})
	if err != nil {
		log.Fatalf("mstl decomposition failed: %v", err)
	}

	mustSave(report, outDir, "mstl_trend", "MSTL Trend", []plot.Series{
		{Label: "observed", Data: hourlySeries, Color: plot.Gold},
		{Label: "trend", Data: mstlResult.Trend, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
	})
	mustSave(report, outDir, "mstl_daily", "MSTL Daily Seasonal Component", []plot.Series{
		{Label: "daily", Data: mstlResult.Seasonal[0].Series, Color: plot.LightSeaGreen, Style: plot.LinePoints},
	})
	mustSave(report, outDir, "mstl_weekly", "MSTL Weekly Seasonal Component", []plot.Series{
		{Label: "weekly", Data: mstlResult.Seasonal[1].Series, Color: plot.Orchid, Style: plot.LinePoints},
	})
	mustSave(report, outDir, "mstl_residual", "MSTL Residual", []plot.Series{
		{Label: "residual", Data: mstlResult.Residual, Color: plot.Crimson, Style: plot.LinePoints},
	})

	if _, err := report.Write(outDir); err != nil {
		log.Fatalf("report generation failed: %v", err)
	}
	exampleutil.PrintOutputDir(outDir)
}

func mustSave(report *exampleutil.Report, outDir string, slug string, title string, series []plot.Series) {
	if err := report.SaveChart(outDir, slug, title, series, plot.TimeFormat("2006")); err != nil {
		log.Fatalf("%s plot failed: %v", slug, err)
	}
}
