package main

import (
	"image/color"
	"log"
	"math"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/anomaly"
	"github.com/wenta/timeseries-go/generator"
	"github.com/wenta/timeseries-go/internal/exampleutil"
	"github.com/wenta/timeseries-go/plot"
)

func main() {
	outDir, err := exampleutil.OutputDir("anomalies")
	if err != nil {
		log.Fatalf("output dir creation failed: %v", err)
	}
	report := exampleutil.NewReport("anomalies", "Anomaly Detection Examples")

	ts, err := syntheticAnomalySeries()
	if err != nil {
		log.Fatalf("synthetic anomaly series creation failed: %v", err)
	}

	zFlags, err := anomaly.FindAnomaliesWithZScore(ts)
	if err != nil {
		log.Fatalf("zscore anomaly detection failed: %v", err)
	}
	robustFlags, err := anomaly.FindAnomaliesWithRobustZScore(ts)
	if err != nil {
		log.Fatalf("robust zscore anomaly detection failed: %v", err)
	}
	spikes, err := anomaly.FindSpikeAnomalies(ts, 10)
	if err != nil {
		log.Fatalf("spike anomaly detection failed: %v", err)
	}
	drops, err := anomaly.FindDropAnomalies(ts, 6)
	if err != nil {
		log.Fatalf("drop anomaly detection failed: %v", err)
	}
	flat, err := anomaly.FindFlatlineAnomalies(ts, 0, 3)
	if err != nil {
		log.Fatalf("flatline anomaly detection failed: %v", err)
	}

	mustSave(report, outDir, "zscore", "Hourly Series with Z-Score Anomalies", ts, exampleutil.FlaggedPoints(ts, zFlags), plot.Crimson)
	mustSave(report, outDir, "robust", "Hourly Series with Robust Z-Score Anomalies", ts, exampleutil.FlaggedPoints(ts, robustFlags), plot.DeepSkyBlue)

	if err := report.SaveChart(
		outDir,
		"specialized",
		"Spike, Drop, and Flatline Markers",
		[]plot.Series{
			{Label: "series", Data: ts, Color: plot.Gold},
			{Label: "spikes", Data: exampleutil.FlaggedPoints(ts, spikes), Color: plot.DarkOrange, Style: plot.Points},
			{Label: "drops", Data: exampleutil.FlaggedPoints(ts, drops), Color: plot.Cyan, Style: plot.Points},
			{Label: "flatline", Data: exampleutil.FlaggedPoints(ts, flat), Color: plot.Chartreuse, Style: plot.Points},
		},
		plot.TimeFormat("2006-01-02"),
	); err != nil {
		log.Fatalf("specialized anomaly plot failed: %v", err)
	}

	if _, err := report.Write(outDir); err != nil {
		log.Fatalf("report generation failed: %v", err)
	}
	exampleutil.PrintOutputDir(outDir)
}

func mustSave(report *exampleutil.Report, outDir string, slug string, title string, base timeseriesgo.TimeSeries, markers timeseriesgo.TimeSeries, clr color.Color) {
	if err := report.SaveChart(
		outDir,
		slug,
		title,
		[]plot.Series{
			{Label: "series", Data: base, Color: plot.Gold},
			{Label: "anomalies", Data: markers, Color: clr, Style: plot.Points},
		},
		plot.TimeFormat("2006-01-02"),
	); err != nil {
		log.Fatalf("%s anomaly plot failed: %v", slug, err)
	}
}

func syntheticAnomalySeries() (timeseriesgo.TimeSeries, error) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	index := generator.MakeSeriesIndex(base, time.Hour, 24*21)
	values := make([]float64, len(index))

	for i := range values {
		daily := 7 * math.Sin(2*math.Pi*float64(i%24)/24)
		weekly := 4 * math.Cos(2*math.Pi*float64(i%168)/168)
		trend := 80.0 + 0.015*float64(i)
		values[i] = trend + daily + weekly
	}

	values[60] += 24
	values[143] -= 22
	values[280] += 18
	for i := 320; i < 330; i++ {
		values[i] = 77.5
	}

	return timeseriesgo.Zip(index, values)
}
