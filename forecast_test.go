package timeseriesgo

import (
	"testing"
	"time"
)

func TestNaiveForecast(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Hour
	count := 5
	index := MakeSeriesIndex(start, interval, count)

	ts := RandomWalk(index, 5)
	forecastHorizon := 3
	forecast := Naive(ts, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Errorf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	lastPoint, err := ts.Last()
	if err != nil {
		t.Errorf("Unexpected error getting last point: %v", err)
	}

	for i, dp := range forecast.datapoints {
		expectedTime := lastPoint.timestamp.Add(time.Duration(i+1) * interval)
		if !dp.timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.timestamp)
		}
		if dp.value != lastPoint.value {
			t.Errorf("At index %d: expected value %f, got %f", i, lastPoint.value, dp.value)
		}
	}
}
