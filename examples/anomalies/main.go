package main

import (
	"image/color"
	"log"
	"path/filepath"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/anomaly"
	"github.com/wenta/timeseries-go/internal/exampleutil"
	"github.com/wenta/timeseries-go/plot"
)

func main() {
	outDir, err := exampleutil.OutputDir("anomalies")
	if err != nil {
		log.Fatalf("output dir creation failed: %v", err)
	}

	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	points := []timeseriesgo.DataPoint{
		{Timestamp: base, Value: 10},
		{Timestamp: base.Add(time.Hour), Value: 11},
		{Timestamp: base.Add(2 * time.Hour), Value: 12},
		{Timestamp: base.Add(3 * time.Hour), Value: 40},
		{Timestamp: base.Add(4 * time.Hour), Value: 13},
		{Timestamp: base.Add(5 * time.Hour), Value: 12},
		{Timestamp: base.Add(6 * time.Hour), Value: 4},
		{Timestamp: base.Add(7 * time.Hour), Value: 4},
		{Timestamp: base.Add(8 * time.Hour), Value: 4},
		{Timestamp: base.Add(9 * time.Hour), Value: 15},
	}
	ts := timeseriesgo.FromDataPoints(points)

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

	mustSave(outDir, "zscore", "Series with Z-Score Anomalies", ts, exampleutil.FlaggedPoints(ts, zFlags), plot.Crimson)
	mustSave(outDir, "robust", "Series with Robust Z-Score Anomalies", ts, exampleutil.FlaggedPoints(ts, robustFlags), plot.DeepSkyBlue)

	if err := exampleutil.SaveAllFormats(
		filepath.Join(outDir, "specialized"),
		[]plot.Series{
			{Label: "series", Data: ts, Color: plot.Gold},
			{Label: "spikes", Data: exampleutil.FlaggedPoints(ts, spikes), Color: plot.DarkOrange, Style: plot.Points},
			{Label: "drops", Data: exampleutil.FlaggedPoints(ts, drops), Color: plot.Cyan, Style: plot.Points},
			{Label: "flatline", Data: exampleutil.FlaggedPoints(ts, flat), Color: plot.Chartreuse, Style: plot.Points},
		},
		plot.Title("Specialized Anomaly Markers"),
	); err != nil {
		log.Fatalf("specialized anomaly plot failed: %v", err)
	}

	exampleutil.PrintOutputDir(outDir)
}

func mustSave(outDir string, slug string, title string, base timeseriesgo.TimeSeries, markers timeseriesgo.TimeSeries, clr color.Color) {
	if err := exampleutil.SaveAllFormats(
		filepath.Join(outDir, slug),
		[]plot.Series{
			{Label: "series", Data: base, Color: plot.Gold},
			{Label: "anomalies", Data: markers, Color: clr, Style: plot.Points},
		},
		plot.Title(title),
	); err != nil {
		log.Fatalf("%s anomaly plot failed: %v", slug, err)
	}
}
