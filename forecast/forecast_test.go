package forecast

import (
	"testing"
	"time"

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
