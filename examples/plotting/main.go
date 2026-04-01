package main

import (
	"log"
	"time"

	"github.com/wenta/timeseries-go/forecast"
	"github.com/wenta/timeseries-go/internal/exampleutil"
	"github.com/wenta/timeseries-go/plot"
	"github.com/wenta/timeseries-go/stats"
)

func main() {
	outDir, err := exampleutil.OutputDir("plotting")
	if err != nil {
		log.Fatalf("output dir creation failed: %v", err)
	}
	report := exampleutil.NewReport("plotting", "Plotting Examples")

	air, err := exampleutil.LoadCSVSeries("examples/data/air_passengers.csv", "2006-01-02", "air-passengers")
	if err != nil {
		log.Fatalf("air passengers csv load failed: %v", err)
	}

	yearlyMA := stats.MovingAverage(air, 365*24*time.Hour)
	forecastSeries := forecast.SimpleExponentialSmoothing(air, 0.2, 12)
	forecastAnchored := exampleutil.AnchoredForecast(air, forecastSeries)

	if err := report.SaveChartSeries(
		outDir,
		"air_passengers",
		"AirPassengers",
		air,
		plot.YLabel("Passengers"),
	); err != nil {
		log.Fatalf("single series plot failed: %v", err)
	}

	if err := report.SaveChart(
		outDir,
		"moving_average",
		"Passengers vs Moving Average",
		[]plot.Series{
			{Label: "passengers", Data: air, Color: plot.Gold},
			{Label: "yearly MA", Data: yearlyMA, Color: plot.LightSeaGreen, Style: plot.Points},
		},
		plot.YLabel("Passengers"),
	); err != nil {
		log.Fatalf("moving average plot failed: %v", err)
	}

	if err := report.SaveChart(
		outDir,
		"ses_forecast",
		"Passengers vs SES Forecast",
		[]plot.Series{
			{Label: "passengers", Data: air, Color: plot.Gold},
			{Label: "ses", Data: forecastAnchored, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
		},
		plot.YLabel("Passengers"),
		plot.TimeFormat("2006"),
	); err != nil {
		log.Fatalf("ses forecast plot failed: %v", err)
	}

	if _, err := report.Write(outDir); err != nil {
		log.Fatalf("report generation failed: %v", err)
	}
	exampleutil.PrintOutputDir(outDir)
}
