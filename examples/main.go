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

	saveChart(outDir, "air_passengers", "AirPassengers", []plot.Series{
		{Label: "air", Data: series, Color: plot.Gold},
	})
	saveChart(outDir, "moving_average", "AirPassengers vs Moving Average", []plot.Series{
		{Label: "air", Data: series, Color: plot.Gold},
		{Label: "moving average", Data: ma, Color: plot.LightSeaGreen, Style: plot.Points},
	})
	saveChart(outDir, "forecast_comparison", "Naive vs SES Forecast", []plot.Series{
		{Label: "air", Data: series, Color: plot.Gold},
		{Label: "naive", Data: fc, Color: plot.DarkOrange, Style: plot.LinePoints},
		{Label: "ses", Data: ses, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
	}, plot.TimeFormat("2006"))

	exampleBase := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	trendIndex := generator.MakeSeriesIndex(exampleBase, 30*time.Minute, 6)
	trendValues := []float64{10, 12, 13, 16, 18, 21}
	trendSeries, err := timeseriesgo.Zip(trendIndex, trendValues)
	if err != nil {
		log.Fatalf("trend series creation failed: %v", err)
	}
	holt := forecast.DoubleExponentialSmoothing(trendSeries, 0.8, 0.2, 3)
	holtEstimated := forecast.DoubleExponentialSmoothingEstimated(trendSeries, 0.8, 0.2, 3)
	saveChart(outDir, "holt_comparison", "Holt vs Holt Estimated", []plot.Series{
		{Label: "trend", Data: trendSeries, Color: plot.Gold},
		{Label: "holt", Data: holt, Color: plot.Cyan, Style: plot.LinePoints},
		{Label: "holt estimated", Data: holtEstimated, Color: plot.Chartreuse, Style: plot.LinePoints},
	})

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

	walk := generator.RandomWalk(generator.MakeSeriesIndex(exampleBase, 30*time.Minute, 24), 0)
	spikeFlags, err := anomaly.FindSpikeAnomalies(walk, 3)
	if err != nil {
		log.Fatalf("spike detection failed: %v", err)
	}
	saveChart(outDir, "spike_anomalies", "Random Walk with Spike Markers", []plot.Series{
		{Label: "walk", Data: walk, Color: plot.MediumPurple},
		{Label: "spikes", Data: exampleutil.FlaggedPoints(walk, spikeFlags), Color: plot.Crimson, Style: plot.Points},
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
