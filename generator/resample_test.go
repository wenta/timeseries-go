package generator

import (
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

func TestBlockBootstrap(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	sourceIndex := MakeSeriesIndex(base, time.Minute, 6)
	sourceValues := []float64{1, 2, 3, 4, 5, 6}
	source, err := timeseriesgo.Zip(sourceIndex, sourceValues)
	if err != nil {
		t.Fatalf("unexpected error creating source: %v", err)
	}

	outputIndex := MakeSeriesIndex(base.Add(time.Hour), time.Minute, 8)
	result := BlockBootstrap(source, outputIndex, 2)

	if result.Length() != len(outputIndex) {
		t.Fatalf("expected output length %d, got %d", len(outputIndex), result.Length())
	}
	for i, dp := range result.DataPoints() {
		if !dp.Timestamp.Equal(outputIndex[i]) {
			t.Fatalf("timestamp at index %d: expected %v, got %v", i, outputIndex[i], dp.Timestamp)
		}
		if !containsFloat(sourceValues, dp.Value) {
			t.Fatalf("value at index %d not found in source values: %f", i, dp.Value)
		}
	}
}

func TestMovingBlockBootstrap(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	sourceIndex := MakeSeriesIndex(base, time.Minute, 5)
	sourceValues := []float64{10, 20, 30, 40, 50}
	source, err := timeseriesgo.Zip(sourceIndex, sourceValues)
	if err != nil {
		t.Fatalf("unexpected error creating source: %v", err)
	}

	outputIndex := MakeSeriesIndex(base.Add(time.Hour), time.Minute, 7)
	result := MovingBlockBootstrap(source, outputIndex, 3)

	if result.Length() != len(outputIndex) {
		t.Fatalf("expected output length %d, got %d", len(outputIndex), result.Length())
	}
	for i, dp := range result.DataPoints() {
		if !dp.Timestamp.Equal(outputIndex[i]) {
			t.Fatalf("timestamp at index %d: expected %v, got %v", i, outputIndex[i], dp.Timestamp)
		}
		if !containsFloat(sourceValues, dp.Value) {
			t.Fatalf("value at index %d not found in source values: %f", i, dp.Value)
		}
	}
}

func TestBootstrapInvalidInput(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := MakeSeriesIndex(base, time.Minute, 3)
	source, err := timeseriesgo.Zip(index, []float64{1, 2, 3})
	if err != nil {
		t.Fatalf("unexpected error creating source: %v", err)
	}

	cases := []timeseriesgo.TimeSeries{
		BlockBootstrap(timeseriesgo.Empty(), index, 2),
		BlockBootstrap(source, nil, 2),
		BlockBootstrap(source, index, 0),
		MovingBlockBootstrap(timeseriesgo.Empty(), index, 2),
		MovingBlockBootstrap(source, nil, 2),
		MovingBlockBootstrap(source, index, -1),
	}

	for i, result := range cases {
		if !result.IsEmpty() {
			t.Fatalf("case %d: expected empty result, got length %d", i, result.Length())
		}
	}
}

func TestSeasonalResampleHourlyBuckets(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	sourceIndex := []time.Time{
		base,
		base.Add(time.Hour),
		base.Add(24 * time.Hour),
		base.Add(25 * time.Hour),
	}
	sourceValues := []float64{10, 20, 10, 20}
	source, err := timeseriesgo.Zip(sourceIndex, sourceValues)
	if err != nil {
		t.Fatalf("unexpected error creating source: %v", err)
	}

	outputIndex := []time.Time{
		base.Add(48 * time.Hour),
		base.Add(49 * time.Hour),
	}

	result := SeasonalResample(source, outputIndex, func(ts time.Time) string {
		return ts.Format("15")
	})

	expected := []float64{10, 20}
	for i, dp := range result.DataPoints() {
		if !dp.Timestamp.Equal(outputIndex[i]) {
			t.Fatalf("timestamp at index %d: expected %v, got %v", i, outputIndex[i], dp.Timestamp)
		}
		if dp.Value != expected[i] {
			t.Fatalf("value at index %d: expected %f, got %f", i, expected[i], dp.Value)
		}
	}
}

func TestSeasonalResampleFallback(t *testing.T) {
	base := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC) // Monday
	sourceIndex := []time.Time{
		base,
		base.Add(24 * time.Hour),
	}
	sourceValues := []float64{42, 42}
	source, err := timeseriesgo.Zip(sourceIndex, sourceValues)
	if err != nil {
		t.Fatalf("unexpected error creating source: %v", err)
	}

	outputIndex := []time.Time{
		base.Add(5 * 24 * time.Hour), // Saturday, missing bucket
		base.Add(6 * 24 * time.Hour), // Sunday, missing bucket
	}

	result := SeasonalResample(source, outputIndex, func(ts time.Time) string {
		return ts.Weekday().String()
	})

	for i, dp := range result.DataPoints() {
		if !dp.Timestamp.Equal(outputIndex[i]) {
			t.Fatalf("timestamp at index %d: expected %v, got %v", i, outputIndex[i], dp.Timestamp)
		}
		if dp.Value != 42 {
			t.Fatalf("value at index %d: expected fallback value 42, got %f", i, dp.Value)
		}
	}
}

func containsFloat(values []float64, target float64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
