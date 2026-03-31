package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/wenta/timeseries-go/forecast"
	"github.com/wenta/timeseries-go/internal/exampledata"
	"github.com/wenta/timeseries-go/internal/exampleutil"
	"github.com/wenta/timeseries-go/plot"
	"github.com/wenta/timeseries-go/stats"
)

func main() {
	outDir, err := exampleutil.OutputDir("plotting")
	if err != nil {
		log.Fatalf("output dir creation failed: %v", err)
	}

	air, err := exampledata.AirPassengers()
	if err != nil {
		log.Fatalf("air passengers series creation failed: %v", err)
	}

	yearlyMA := stats.MovingAverage(air, 365*24*time.Hour)
	forecastSeries := forecast.SimpleExponentialSmoothing(air, 0.2, 12)
	forecastAnchored := exampleutil.AnchoredForecast(air, forecastSeries)

	if err := exampleutil.SaveAllFormatsSeries(
		filepath.Join(outDir, "air_passengers"),
		air,
		plot.Title("AirPassengers"),
		plot.YLabel("Passengers"),
	); err != nil {
		log.Fatalf("single series plot failed: %v", err)
	}

	if err := exampleutil.SaveAllFormats(
		filepath.Join(outDir, "moving_average"),
		[]plot.Series{
			{Label: "passengers", Data: air, Color: plot.Gold},
			{Label: "yearly MA", Data: yearlyMA, Color: plot.LightSeaGreen, Style: plot.Points},
		},
		plot.Title("Passengers vs Moving Average"),
		plot.YLabel("Passengers"),
	); err != nil {
		log.Fatalf("moving average plot failed: %v", err)
	}

	if err := exampleutil.SaveAllFormats(
		filepath.Join(outDir, "ses_forecast"),
		[]plot.Series{
			{Label: "passengers", Data: air, Color: plot.Gold},
			{Label: "ses", Data: forecastAnchored, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
		},
		plot.Title("Passengers vs SES Forecast"),
		plot.YLabel("Passengers"),
		plot.TimeFormat("2006"),
	); err != nil {
		log.Fatalf("ses forecast plot failed: %v", err)
	}

	exampleutil.PrintOutputDir(outDir)
}
