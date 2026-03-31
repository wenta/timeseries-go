package main

import (
	"log"
	"path/filepath"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/internal/exampleutil"
	"github.com/wenta/timeseries-go/plot"
)

func main() {
	outDir, err := exampleutil.OutputDir("transformations")
	if err != nil {
		log.Fatalf("output dir creation failed: %v", err)
	}

	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	points := []timeseriesgo.DataPoint{
		{Timestamp: base, Value: 10},
		{Timestamp: base.Add(2 * time.Hour), Value: 16},
		{Timestamp: base.Add(5 * time.Hour), Value: 13},
		{Timestamp: base.Add(7 * time.Hour), Value: 20},
	}
	ts := timeseriesgo.FromDataPoints(points)

	sliced := ts.Slice(base.Add(time.Hour), base.Add(6*time.Hour))
	scaled := ts.MapValues(func(v float64) float64 { return v * 1.2 })
	highOnly := scaled.Filter(func(dp timeseriesgo.DataPoint) bool { return dp.Value >= 18 })
	resampled := ts.ResampleWithDefaultValue(time.Hour, -1)
	interpolated := ts.Interpolate(time.Hour)
	diffs := interpolated.Differentiate()
	pairwiseSums := interpolated.Integrate()

	mustSave(outDir, "original_sparse", "Original Sparse Series", []plot.Series{
		{Label: "original", Data: ts, Color: plot.Gold},
	})
	mustSave(outDir, "slice", "Slice", []plot.Series{
		{Label: "original", Data: ts, Color: plot.Gold},
		{Label: "slice", Data: sliced, Color: plot.Cyan, Style: plot.Points},
	})
	mustSave(outDir, "filter", "Filter >= 18", []plot.Series{
		{Label: "scaled", Data: scaled, Color: plot.LightSeaGreen},
		{Label: "filter >= 18", Data: highOnly, Color: plot.Crimson, Style: plot.Points},
	})
	mustSave(outDir, "resample_variants", "Resample Variants", []plot.Series{
		{Label: "original", Data: ts, Color: plot.Gold, Style: plot.Points},
		{Label: "interpolated", Data: interpolated, Color: plot.DeepSkyBlue},
		{Label: "default fill", Data: resampled, Color: plot.Orchid, Style: plot.LinePoints},
	})
	mustSave(outDir, "scaled", "Scaled", []plot.Series{
		{Label: "original", Data: ts, Color: plot.Gold},
		{Label: "scaled", Data: scaled, Color: plot.LightSeaGreen},
	})
	mustSave(outDir, "differentiate", "Differentiate", []plot.Series{
		{Label: "interpolated", Data: interpolated, Color: plot.DeepSkyBlue},
		{Label: "differentiate", Data: diffs, Color: plot.DarkOrange},
	})
	mustSave(outDir, "integrate", "Pairwise Sums (Integrate)", []plot.Series{
		{Label: "interpolated input", Data: interpolated, Color: plot.DeepSkyBlue},
		{Label: "pairwise sums", Data: pairwiseSums, Color: plot.MediumPurple},
	})

	exampleutil.PrintOutputDir(outDir)
}

func mustSave(outDir string, slug string, title string, series []plot.Series) {
	if err := exampleutil.SaveAllFormats(filepath.Join(outDir, slug), series, plot.Title(title)); err != nil {
		log.Fatalf("%s plot failed: %v", slug, err)
	}
}
