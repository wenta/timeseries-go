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

func TestRepeat(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	pattern := Empty()
	pattern.AddPoint(DataPoint{timestamp: base, value: 1})
	pattern.AddPoint(DataPoint{timestamp: base.Add(time.Minute), value: 2})

	start := base
	end := base.Add(5 * time.Minute)

	repeated := Repeat(pattern, start, end)

	if repeated.Length() != 5 {
		t.Fatalf("Expected repeated length 5, got %d", repeated.Length())
	}

	expectedValues := []float64{1, 2, 1, 2, 1}
	for i, dp := range repeated.datapoints {
		expectedTs := start.Add(time.Duration(i) * time.Minute)
		if !dp.timestamp.Equal(expectedTs) {
			t.Errorf("At idx %d expected timestamp %v, got %v", i, expectedTs, dp.timestamp)
		}
		if dp.value != expectedValues[i] {
			t.Errorf("At idx %d expected value %.0f, got %.0f", i, expectedValues[i], dp.value)
		}
	}
}

func TestRepeatSinglePointPattern(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	pattern := Empty()
	pattern.AddPoint(DataPoint{timestamp: base, value: 5})

	repeated := Repeat(pattern, base, base.Add(10*time.Minute))

	if repeated.Length() != 1 {
		t.Fatalf("Expected pattern returned unchanged with length 1, got %d", repeated.Length())
	}
	if repeated.datapoints[0].timestamp != base || repeated.datapoints[0].value != 5 {
		t.Errorf("Expected original datapoint preserved, got %+v", repeated.datapoints[0])
	}
}
