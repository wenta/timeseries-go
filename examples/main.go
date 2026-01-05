package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"strings"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/anomaly"
	"github.com/wenta/timeseries-go/forecast"
	"github.com/wenta/timeseries-go/generator"
	"github.com/wenta/timeseries-go/metrics"
	"github.com/wenta/timeseries-go/stats"
	"github.com/wenta/timeseries-go/tsio"
)

func main() {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := generator.MakeSeriesIndex(base, 30*time.Minute, 10)

	// 1) Generate a simple series and compute moving average.
	series := generator.RandomWalk(index, 10)
	series.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(5 * time.Hour), Value: 30})
	fmt.Println("Original series values:", series.Values())

	ma := stats.MovingAverage(series, time.Hour)
	fmt.Println("Hourly moving average:", ma.Values())

	// 2) Naive forecast for 3 future points.
	fc := forecast.Naive(series, 3)
	fmt.Println("Naive forecast values:", fc.Values())

	// 3) Simple exponential smoothing forecast.
	ses := forecast.SimpleExponentialSmoothing(series, 0.2, 3)
	fmt.Println("SES forecast values:", ses.Values())

	// 4) Z-Score anomalies on the original series.
	flags, err := anomaly.FindAnomaliesWithZScore(series)
	if err != nil {
		log.Fatalf("zscore failed: %v", err)
	}
	fmt.Println("Z-Score anomaly flags:", flags.Values())

	// 5) Serialize to CSV and back.
	csvStr, err := tsio.ToString(series)
	if err != nil {
		log.Fatalf("serialize failed: %v", err)
	}
	fmt.Println("\nCSV output:\n", csvStr)

	reloaded, err := tsio.FromString(*csv.NewReader(strings.NewReader(csvStr)), "example")
	if err != nil {
		log.Fatalf("parse failed: %v", err)
	}
	fmt.Printf("Reloaded length: %d\n", reloaded.Length())

	// 6) Compare two series: MSE/RMSE/MAE.
	index2 := generator.MakeSeriesIndex(base, 30*time.Minute, 10)
	series2 := generator.Constant(index2, 9)
	mse, _ := metrics.MSE(series, series2)
	rmse, _ := metrics.RMSE(series, series2)
	mae, _ := metrics.MAE(series, series2)
	fmt.Printf("MSE=%.2f RMSE=%.2f MAE=%.2f\n", mse, rmse, mae)

	// 7) Detect spikes in a random walk.
	walk := generator.RandomWalk(index, 0)
	spikeFlags, err := anomaly.FindSpikeAnomalies(walk, 3)
	if err != nil {
		log.Fatalf("spike detection failed: %v", err)
	}
	fmt.Println("Random walk spike flags:", spikeFlags.Values())

	// 8) Merge forecast with original and compute MAD on spike flags.
	merged := series.Merge(fc)
	fmt.Println("Merged series length:", merged.Length())
	mad, _ := metrics.MAD(spikeFlags)
	fmt.Printf("MAD of spike flags: %.2f\n", mad)
}
