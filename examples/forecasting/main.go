package main

import (
	"log"
	"path/filepath"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/forecast"
	"github.com/wenta/timeseries-go/internal/exampledata"
	"github.com/wenta/timeseries-go/internal/exampleutil"
	"github.com/wenta/timeseries-go/plot"
)

func main() {
	outDir, err := exampleutil.OutputDir("forecasting")
	if err != nil {
		log.Fatalf("output dir creation failed: %v", err)
	}

	air, err := exampledata.AirPassengers()
	if err != nil {
		log.Fatalf("air passengers series creation failed: %v", err)
	}

	naive := forecast.Naive(air, 12)
	ses := forecast.SimpleExponentialSmoothing(air, 0.2, 12)

	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	trendIndex := []time.Time{
		base,
		base.Add(30 * time.Minute),
		base.Add(60 * time.Minute),
		base.Add(90 * time.Minute),
		base.Add(120 * time.Minute),
		base.Add(150 * time.Minute),
	}
	trendValues := []float64{10, 12, 13, 16, 18, 21}
	trendSeries, err := timeseriesgo.Zip(trendIndex, trendValues)
	if err != nil {
		log.Fatalf("trend series creation failed: %v", err)
	}

	holt := forecast.DoubleExponentialSmoothing(trendSeries, 0.8, 0.2, 3)
	holtEstimated := forecast.DoubleExponentialSmoothingEstimated(trendSeries, 0.8, 0.2, 3)

	saveChart(outDir, "air_forecast_comparison", "AirPassengers Forecast Comparison", []plot.Series{
		{Label: "actual", Data: air, Color: plot.Gold},
		{Label: "naive", Data: naive, Color: plot.DarkOrange, Style: plot.LinePoints},
		{Label: "ses", Data: ses, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
	}, plot.TimeFormat("2006"))

	saveChart(outDir, "holt_comparison", "Holt Initialization Comparison", []plot.Series{
		{Label: "trend", Data: trendSeries, Color: plot.Gold},
		{Label: "holt", Data: holt, Color: plot.Cyan, Style: plot.LinePoints},
		{Label: "holt estimated", Data: holtEstimated, Color: plot.Chartreuse, Style: plot.LinePoints},
	})

	exampleutil.PrintOutputDir(outDir)
}

func saveChart(outDir string, slug string, title string, series []plot.Series, opts ...plot.Option) {
	if err := exampleutil.SaveAllFormats(
		filepath.Join(outDir, slug),
		series,
		append([]plot.Option{plot.Title(title)}, opts...)...,
	); err != nil {
		log.Fatalf("%s plot failed: %v", slug, err)
	}
}
