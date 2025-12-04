package timeseriesgo

import (
	"testing"
	"time"
)

func TestGenerateConstant(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Hour
	count := 5
	value := 42.0
	index := MakeSeriesIndex(start, interval, count)

	ts := Constant(index, value)

	if ts.Length() != count {
		t.Errorf("Expected TimeSeries length %d, got %d", count, ts.Length())
	}

	for i, v := range ts.Values() {
		if v != value {
			t.Errorf("At index %d: expected value %f, got %f", i, value, v)
		}
		expectedTime := start.Add(time.Duration(i) * interval)
		if !ts.datapoints[i].timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, ts.datapoints[i].timestamp)
		}
	}
}
func TestGenerateConstant_Empty(t *testing.T) {
	index := MakeSeriesIndex(time.Now(), time.Minute, 0)
	ts := Constant(index, 100.0)
	if !ts.IsEmpty() {
		t.Errorf("Expected empty TimeSeries, got length %d", ts.Length())
	}
}

func TestGenerateRandomWalk(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Minute
	count := 10
	startValue := 50.0
	index := MakeSeriesIndex(start, interval, count)

	ts := RandomWalk(index, startValue)

	if ts.Length() != count {
		t.Errorf("Expected TimeSeries length %d, got %d", count, ts.Length())
	}

}
