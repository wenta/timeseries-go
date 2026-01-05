package forecast

import (
	"math"
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/generator"
)

func TestNaiveForecast(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Hour
	count := 5
	index := generator.MakeSeriesIndex(start, interval, count)

	ts := generator.RandomWalk(index, 5)
	forecastHorizon := 3
	forecast := Naive(ts, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Errorf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	lastPoint, err := ts.Last()
	if err != nil {
		t.Errorf("Unexpected error getting last point: %v", err)
	}

	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if dp.Value != lastPoint.Value {
			t.Errorf("At index %d: expected value %f, got %f", i, lastPoint.Value, dp.Value)
		}
	}
}

func TestSimpleExponentialSmoothing(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Hour
	values := []float64{10, 12, 13, 12}
	index := generator.MakeSeriesIndex(start, interval, len(values))

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	forecastHorizon := 3
	alpha := 0.5
	forecast := SimpleExponentialSmoothing(ts, alpha, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Errorf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	expectedValue := 12.0
	lastPoint, err := ts.Last()
	if err != nil {
		t.Fatalf("Unexpected error getting last point: %v", err)
	}
	forecast.Print()
	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if dp.Value != expectedValue {
			t.Errorf("At index %d: expected value %f, got %f", i, expectedValue, dp.Value)
		}
	}
}

func TestSimpleExponentialSmoothingStatsmodelsOilData(t *testing.T) {
	// Data from the statsmodels exponential smoothing example (Saudi Arabia oil).
	values := []float64{
		446.6565,
		454.4733,
		455.663,
		423.6322,
		456.2713,
		440.5881,
		425.3325,
		485.1494,
		506.0482,
		526.792,
		514.2689,
		494.211,
	}
	start := time.Date(1996, 12, 31, 0, 0, 0, 0, time.UTC)
	interval := 365 * 24 * time.Hour
	index := generator.MakeSeriesIndex(start, interval, len(values))

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	forecastHorizon := 3
	alpha := 0.2
	forecast := SimpleExponentialSmoothing(ts, alpha, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Errorf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	expectedValue := 484.80246538161776
	const epsilon = 1e-9
	lastPoint, err := ts.Last()
	if err != nil {
		t.Fatalf("Unexpected error getting last point: %v", err)
	}

	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if math.Abs(dp.Value-expectedValue) > epsilon {
			t.Errorf("At index %d: expected value %f, got %f", i, expectedValue, dp.Value)
		}
	}
}
