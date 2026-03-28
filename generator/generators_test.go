package generator

import (
	"math"
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
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
		if !ts.DataPoints()[i].Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, ts.DataPoints()[i].Timestamp)
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

func TestGenerateRandomNoiseZeroStddev(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Minute
	count := 10
	mean := 42.5
	index := MakeSeriesIndex(start, interval, count)

	ts := RandomNoise(index, mean, 0)

	if ts.Length() != count {
		t.Fatalf("Expected TimeSeries length %d, got %d", count, ts.Length())
	}

	for i, dp := range ts.DataPoints() {
		expectedTime := start.Add(time.Duration(i) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if dp.Value != mean {
			t.Errorf("At index %d: expected value %f, got %f", i, mean, dp.Value)
		}
	}
}

func TestGenerateRandomNoiseApproximateMoments(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	count := 10000
	mean := 5.0
	stddev := 2.0
	index := MakeSeriesIndex(start, time.Second, count)

	ts := RandomNoise(index, mean, stddev)

	if ts.Length() != count {
		t.Fatalf("Expected TimeSeries length %d, got %d", count, ts.Length())
	}

	values := ts.Values()
	var sum float64
	for _, v := range values {
		sum += v
	}
	sampleMean := sum / float64(len(values))

	var squaredDiffs float64
	for _, v := range values {
		diff := v - sampleMean
		squaredDiffs += diff * diff
	}
	sampleStddev := math.Sqrt(squaredDiffs / float64(len(values)-1))

	if math.Abs(sampleMean-mean) > 0.1 {
		t.Errorf("Expected sample mean near %f, got %f", mean, sampleMean)
	}
	if math.Abs(sampleStddev-stddev) > 0.1 {
		t.Errorf("Expected sample stddev near %f, got %f", stddev, sampleStddev)
	}
}

func TestRepeat(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	pattern := timeseriesgo.Empty()
	pattern.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 1})
	pattern.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(time.Minute), Value: 2})

	start := base
	end := base.Add(5 * time.Minute)

	repeated := Repeat(pattern, start, end)

	if repeated.Length() != 5 {
		t.Fatalf("Expected repeated length 5, got %d", repeated.Length())
	}

	expectedValues := []float64{1, 2, 1, 2, 1}
	for i, dp := range repeated.DataPoints() {
		expectedTs := start.Add(time.Duration(i) * time.Minute)
		if !dp.Timestamp.Equal(expectedTs) {
			t.Errorf("At idx %d expected timestamp %v, got %v", i, expectedTs, dp.Timestamp)
		}
		if dp.Value != expectedValues[i] {
			t.Errorf("At idx %d expected value %.0f, got %.0f", i, expectedValues[i], dp.Value)
		}
	}
}

func TestRepeatSinglePointPattern(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	pattern := timeseriesgo.Empty()
	pattern.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 5})

	repeated := Repeat(pattern, base, base.Add(10*time.Minute))

	if repeated.Length() != 1 {
		t.Fatalf("Expected pattern returned unchanged with length 1, got %d", repeated.Length())
	}
	points := repeated.DataPoints()
	if points[0].Timestamp != base || points[0].Value != 5 {
		t.Errorf("Expected original datapoint preserved, got %+v", points[0])
	}
}
