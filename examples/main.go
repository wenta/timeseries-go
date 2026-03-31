package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/anomaly"
	"github.com/wenta/timeseries-go/forecast"
	"github.com/wenta/timeseries-go/generator"
	"github.com/wenta/timeseries-go/internal/exampledata"
	"github.com/wenta/timeseries-go/internal/exampleutil"
	"github.com/wenta/timeseries-go/metrics"
	"github.com/wenta/timeseries-go/plot"
	"github.com/wenta/timeseries-go/stats"
	"github.com/wenta/timeseries-go/tsio"
)

func main() {
	outDir, err := exampleutil.OutputDir("main")
	if err != nil {
		log.Fatalf("output dir creation failed: %v", err)
	}

	series, err := exampledata.AirPassengers()
	if err != nil {
		log.Fatalf("air passengers series creation failed: %v", err)
	}

	fmt.Printf("AirPassengers length: %d\n", series.Length())
	fmt.Println("AirPassengers first 12 values:", previewValues(series.Values(), 12))

	ma := stats.MovingAverage(series, 365*24*time.Hour)
	fc := forecast.Naive(series, 12)
	ses := forecast.SimpleExponentialSmoothing(series, 0.2, 12)
	fcAnchored := exampleutil.AnchoredForecast(series, fc)
	sesAnchored := exampleutil.AnchoredForecast(series, ses)

	saveChart(outDir, "air_passengers", "AirPassengers", []plot.Series{
		{Label: "air", Data: series, Color: plot.Gold},
	})
	saveChart(outDir, "moving_average", "AirPassengers vs Moving Average", []plot.Series{
		{Label: "air", Data: series, Color: plot.Gold},
		{Label: "moving average", Data: ma, Color: plot.LightSeaGreen, Style: plot.Points},
	})
	saveChart(outDir, "forecast_comparison", "Naive vs SES Forecast", []plot.Series{
		{Label: "air", Data: series, Color: plot.Gold},
		{Label: "naive", Data: fcAnchored, Color: plot.DarkOrange, Style: plot.LinePoints},
		{Label: "ses", Data: sesAnchored, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
	}, plot.TimeFormat("2006"))

	exampleBase := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	trendIndex := []time.Time{
		exampleBase,
		exampleBase.AddDate(0, 1, 0),
		exampleBase.AddDate(0, 2, 0),
		exampleBase.AddDate(0, 3, 0),
		exampleBase.AddDate(0, 4, 0),
		exampleBase.AddDate(0, 5, 0),
		exampleBase.AddDate(0, 6, 0),
		exampleBase.AddDate(0, 7, 0),
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
	saveChart(outDir, "holt_comparison", "Holt vs Holt Estimated", []plot.Series{
		{Label: "trend", Data: trendSeries, Color: plot.Gold},
		{Label: "holt", Data: holtAnchored, Color: plot.Cyan, Style: plot.LinePoints},
		{Label: "holt estimated", Data: holtEstimatedAnchored, Color: plot.Chartreuse, Style: plot.LinePoints},
	}, plot.TimeFormat("2006-01"))

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
	saveChart(outDir, "holt_winters_additive", "Holt-Winters Additive Seasonal Forecast", []plot.Series{
		{Label: "seasonal", Data: seasonalSeries, Color: plot.Gold},
		{Label: "holt-winters", Data: holtWintersAnchored, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
	}, plot.TimeFormat("2006-01"))

	flags, err := anomaly.FindAnomaliesWithZScore(series)
	if err != nil {
		log.Fatalf("zscore failed: %v", err)
	}
	saveChart(outDir, "zscore_anomalies", "Z-Score Anomalies", []plot.Series{
		{Label: "air", Data: series, Color: plot.Gold},
		{Label: "anomalies", Data: exampleutil.FlaggedPoints(series, flags), Color: plot.Crimson, Style: plot.Points},
	})

	csvStr, err := tsio.ToString(series)
	if err != nil {
		log.Fatalf("serialize failed: %v", err)
	}
	fmt.Println("CSV preview:\n" + previewLines(csvStr, 6))

	reloaded, err := tsio.FromString(*csv.NewReader(strings.NewReader(csvStr)), "example")
	if err != nil {
		log.Fatalf("parse failed: %v", err)
	}
	saveChart(outDir, "reloaded", "Original vs Reloaded", []plot.Series{
		{Label: "original", Data: series, Color: plot.Gold},
		{Label: "reloaded", Data: reloaded, Color: plot.Orchid},
	})

	series2 := generator.Constant(series.Timestamps(), 300)
	mse, _ := metrics.MSE(series, series2)
	rmse, _ := metrics.RMSE(series, series2)
	mae, _ := metrics.MAE(series, series2)
	fmt.Printf("MSE=%.2f RMSE=%.2f MAE=%.2f\n", mse, rmse, mae)
	saveChart(outDir, "constant_baseline", "AirPassengers vs Constant Baseline", []plot.Series{
		{Label: "air", Data: series, Color: plot.Gold},
		{Label: "baseline", Data: series2, Color: plot.MediumPurple},
	})

	spikeIndex := generator.MakeSeriesIndex(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), 30*time.Minute, 12)
	spikeValues := []float64{10, 11, 10, 12, 11, 26, 12, 11, 10, 24, 11, 10}
	spikeSeries, err := timeseriesgo.Zip(spikeIndex, spikeValues)
	if err != nil {
		log.Fatalf("spike series creation failed: %v", err)
	}
	spikeFlags, err := anomaly.FindSpikeAnomalies(spikeSeries, 8)
	if err != nil {
		log.Fatalf("spike detection failed: %v", err)
	}
	saveChart(outDir, "spike_anomalies", "Series with Spike Markers", []plot.Series{
		{Label: "series", Data: spikeSeries, Color: plot.MediumPurple},
		{Label: "spikes", Data: exampleutil.FlaggedPoints(spikeSeries, spikeFlags), Color: plot.Crimson, Style: plot.Points},
	})

	mad, _ := metrics.MAD(spikeFlags)
	fmt.Printf("MAD of spike flags: %.2f\n", mad)
	exampleutil.PrintOutputDir(outDir)
}

func saveChart(outDir string, slug string, title string, series []plot.Series, opts ...plot.Option) {
	if err := exampleutil.SaveAllFormats(filepath.Join(outDir, slug), series, append([]plot.Option{plot.Title(title)}, opts...)...); err != nil {
		log.Fatalf("%s plot failed: %v", slug, err)
	}
}

func previewValues(values []float64, limit int) []float64 {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func previewLines(text string, limit int) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) <= limit {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[:limit], "\n") + "\n..."
}
