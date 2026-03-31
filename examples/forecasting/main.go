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
	naiveAnchored := exampleutil.AnchoredForecast(air, naive)
	sesAnchored := exampleutil.AnchoredForecast(air, ses)

	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	trendIndex := []time.Time{
		base,
		base.AddDate(0, 1, 0),
		base.AddDate(0, 2, 0),
		base.AddDate(0, 3, 0),
		base.AddDate(0, 4, 0),
		base.AddDate(0, 5, 0),
		base.AddDate(0, 6, 0),
		base.AddDate(0, 7, 0),
	}
	trendValues := []float64{120, 128, 133, 142, 150, 159, 166, 176}
	trendSeries, err := timeseriesgo.Zip(trendIndex, trendValues)
	if err != nil {
		log.Fatalf("trend series creation failed: %v", err)
	}

	holt := forecast.DoubleExponentialSmoothing(trendSeries, 0.8, 0.2, 3)
	holtEstimated := forecast.DoubleExponentialSmoothingEstimated(trendSeries, 0.8, 0.2, 3)
	holtAnchored := exampleutil.AnchoredForecast(trendSeries, holt)
	holtEstimatedAnchored := exampleutil.AnchoredForecast(trendSeries, holtEstimated)

	seasonalBase := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	seasonalIndex := []time.Time{
		seasonalBase,
		seasonalBase.AddDate(0, 3, 0),
		seasonalBase.AddDate(0, 6, 0),
		seasonalBase.AddDate(0, 9, 0),
		seasonalBase.AddDate(1, 0, 0),
		seasonalBase.AddDate(1, 3, 0),
		seasonalBase.AddDate(1, 6, 0),
		seasonalBase.AddDate(1, 9, 0),
		seasonalBase.AddDate(2, 0, 0),
		seasonalBase.AddDate(2, 3, 0),
		seasonalBase.AddDate(2, 6, 0),
		seasonalBase.AddDate(2, 9, 0),
		seasonalBase.AddDate(3, 0, 0),
		seasonalBase.AddDate(3, 3, 0),
		seasonalBase.AddDate(3, 6, 0),
		seasonalBase.AddDate(3, 9, 0),
	}
	seasonalValues := []float64{
		80, 102, 91, 114,
		86, 108, 97, 120,
		92, 114, 103, 126,
		98, 120, 109, 132,
	}
	seasonalSeries, err := timeseriesgo.Zip(seasonalIndex, seasonalValues)
	if err != nil {
		log.Fatalf("seasonal series creation failed: %v", err)
	}
	holtWinters := forecast.TripleExponentialSmoothing(seasonalSeries, 0.6, 0.3, 0.2, 4, 4)
	holtWintersAnchored := exampleutil.AnchoredForecast(seasonalSeries, holtWinters)

	saveChart(outDir, "air_forecast_comparison", "AirPassengers Forecast Comparison", []plot.Series{
		{Label: "actual", Data: air, Color: plot.Gold},
		{Label: "naive", Data: naiveAnchored, Color: plot.DarkOrange, Style: plot.LinePoints},
		{Label: "ses", Data: sesAnchored, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
	}, plot.TimeFormat("2006"))

	saveChart(outDir, "holt_comparison", "Holt Initialization Comparison", []plot.Series{
		{Label: "trend", Data: trendSeries, Color: plot.Gold},
		{Label: "holt", Data: holtAnchored, Color: plot.Cyan, Style: plot.LinePoints},
		{Label: "holt estimated", Data: holtEstimatedAnchored, Color: plot.Chartreuse, Style: plot.LinePoints},
	}, plot.TimeFormat("2006-01"))

	saveChart(outDir, "holt_winters_additive", "Holt-Winters Additive Seasonal Forecast", []plot.Series{
		{Label: "seasonal", Data: seasonalSeries, Color: plot.Gold},
		{Label: "holt-winters", Data: holtWintersAnchored, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
	}, plot.TimeFormat("2006-01"))

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
