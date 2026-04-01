package main

import (
	"log"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/generator"
	"github.com/wenta/timeseries-go/internal/exampleutil"
	"github.com/wenta/timeseries-go/plot"
)

func main() {
	outDir, err := exampleutil.OutputDir("generators")
	if err != nil {
		log.Fatalf("output dir creation failed: %v", err)
	}
	report := exampleutil.NewReport("generators", "Generator Examples")

	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := generator.MakeSeriesIndex(base, time.Hour, 24*14)

	constant := generator.Constant(index, 5)
	walk := generator.RandomWalk(index, 10)
	noise := generator.RandomNoise(index, 0, 1)
	uniform := generator.UniformNoise(index, -1, 1)
	pulses := generator.PulseTrain(index, []generator.Pulse{
		{StartIndex: 18, Duration: 6, Amplitude: 6},
		{StartIndex: 67, Duration: 5, Amplitude: 4},
		{StartIndex: 124, Duration: 8, Amplitude: 3},
		{StartIndex: 205, Duration: 10, Amplitude: 5},
		{StartIndex: 286, Duration: 4, Amplitude: 7},
	})

	patternIndex := generator.MakeSeriesIndex(base, time.Hour, 24)
	patternValues := make([]float64, len(patternIndex))
	for i := range patternValues {
		hour := i % 24
		switch {
		case hour >= 0 && hour < 5:
			patternValues[i] = 1.2
		case hour >= 6 && hour < 10:
			patternValues[i] = 4.0
		case hour >= 10 && hour < 17:
			patternValues[i] = 2.4
		case hour >= 17 && hour < 22:
			patternValues[i] = 5.1
		default:
			patternValues[i] = 2.0
		}
	}
	pattern, err := timeseriesgo.Zip(patternIndex, patternValues)
	if err != nil {
		log.Fatalf("pattern creation failed: %v", err)
	}
	repeated := generator.Repeat(pattern, base, base.Add(14*24*time.Hour))
	bootstrapIndex := generator.MakeSeriesIndex(base.Add(14*24*time.Hour), time.Hour, 24*7)
	bootstrapped := generator.MovingBlockBootstrap(walk, bootstrapIndex, 12)
	seasonalResampled := generator.SeasonalResample(repeated, bootstrapIndex, func(ts time.Time) string {
		return ts.Format("15:04")
	})

	knnSourceIndex := generator.MakeSeriesIndex(base, time.Hour, 24*10)
	knnSourceValues := make([]float64, len(knnSourceIndex))
	for i := range knnSourceValues {
		hour := i % 24
		daily := 0.0
		switch {
		case hour >= 6 && hour < 10:
			daily = 3.2
		case hour >= 10 && hour < 17:
			daily = 1.1
		case hour >= 17 && hour < 22:
			daily = 4.0
		default:
			daily = 0.4
		}
		weekly := []float64{-0.5, -0.2, 0, 0.2, 0.5, 1.1, 0.8}[(i/24)%7]
		knnSourceValues[i] = 8 + daily + weekly
	}
	knnSource, err := timeseriesgo.Zip(knnSourceIndex, knnSourceValues)
	if err != nil {
		log.Fatalf("knn source creation failed: %v", err)
	}
	knnIndex := generator.MakeSeriesIndex(base.Add(10*24*time.Hour), time.Hour, 24*5)
	knnResampled := generator.ResampleKNN(knnSource, knnIndex, 24, 3)

	endUseIndex := generator.MakeSeriesIndex(base, 15*time.Minute, 24*4*7)
	endUseEvents := generator.EndUseEvents(base, base.Add(7*24*time.Hour), generator.EndUseConfig{
		EventsPerDayMean: 5,
		DurationMean:     12 * time.Minute,
		IntensityMean:    0.18,
	})
	renderedEvents := generator.RenderEvents(endUseIndex, endUseEvents)

	householdIndex := generator.MakeSeriesIndex(base, 15*time.Minute, 24*4*14)
	household := generator.HouseholdDemand(householdIndex, generator.HouseholdDemandConfig{Occupants: 3})

	mustSave(report, outDir, "constant", "Constant Generator", []plot.Series{
		{Label: "constant", Data: constant, Color: plot.LightSeaGreen},
	}, plot.TimeFormat("2006-01-02"))
	mustSave(report, outDir, "constant_vs_walk", "Constant vs Random Walk", []plot.Series{
		{Label: "constant", Data: constant, Color: plot.LightSeaGreen},
		{Label: "walk", Data: walk, Color: plot.MediumPurple},
	}, plot.TimeFormat("2006-01-02"))
	mustSave(report, outDir, "constant_vs_noise", "Constant vs Random Noise", []plot.Series{
		{Label: "constant", Data: constant, Color: plot.LightSeaGreen},
		{Label: "noise", Data: noise, Color: plot.Orchid},
	}, plot.TimeFormat("2006-01-02"))
	mustSave(report, outDir, "noise_comparison", "Gaussian vs Uniform Noise", []plot.Series{
		{Label: "gaussian", Data: noise, Color: plot.Orchid},
		{Label: "uniform", Data: uniform, Color: plot.DarkOrange},
	}, plot.TimeFormat("2006-01-02"))
	mustSave(report, outDir, "pulse_train", "Pulse Train", []plot.Series{
		{Label: "pulses", Data: pulses, Color: plot.Crimson, Style: plot.LinePoints},
	}, plot.TimeFormat("2006-01-02"))
	mustSave(report, outDir, "pattern_vs_repeat", "Pattern vs Repeat", []plot.Series{
		{Label: "pattern", Data: pattern, Color: plot.Cyan, Style: plot.LinePoints},
		{Label: "repeat", Data: repeated, Color: plot.Gold},
	}, plot.TimeFormat("2006-01-02"))
	mustSave(report, outDir, "bootstrap_vs_resample", "Bootstrap vs Seasonal Resample", []plot.Series{
		{Label: "bootstrap", Data: bootstrapped, Color: plot.MediumPurple},
		{Label: "seasonal resample", Data: seasonalResampled, Color: plot.DeepSkyBlue},
	}, plot.TimeFormat("2006-01-02"))
	mustSave(report, outDir, "knn_resample", "KNN Resample", []plot.Series{
		{Label: "source", Data: knnSource, Color: plot.Gold},
		{Label: "knn", Data: exampleutil.AnchoredForecast(knnSource, knnResampled), Color: plot.DeepSkyBlue},
	}, plot.TimeFormat("2006-01-02"))
	mustSave(report, outDir, "rendered_events", "Rendered End-Use Events", []plot.Series{
		{Label: "end-use demand", Data: renderedEvents, Color: plot.Crimson, Style: plot.LinePoints},
	}, plot.TimeFormat("2006-01-02"))
	mustSave(report, outDir, "household_demand", "Household Demand", []plot.Series{
		{Label: "household", Data: household, Color: plot.MediumPurple},
	}, plot.TimeFormat("2006-01-02"))

	if _, err := report.Write(outDir); err != nil {
		log.Fatalf("report generation failed: %v", err)
	}
	exampleutil.PrintOutputDir(outDir)
}

func mustSave(report *exampleutil.Report, outDir string, slug string, title string, series []plot.Series, opts ...plot.Option) {
	if err := report.SaveChart(outDir, slug, title, series, opts...); err != nil {
		log.Fatalf("%s plot failed: %v", slug, err)
	}
}
