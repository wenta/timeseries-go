package generator

import (
	"testing"
	"time"
)

func TestUniformNoiseEmptyIndex(t *testing.T) {
	if result := UniformNoise(nil, -1, 1); !result.IsEmpty() {
		t.Fatalf("expected empty series, got length %d", result.Length())
	}
}

func TestUniformNoiseConstant(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := MakeSeriesIndex(base, time.Minute, 5)

	ts := UniformNoise(index, 3.5, 3.5)
	for i, dp := range ts.DataPoints() {
		if !dp.Timestamp.Equal(index[i]) {
			t.Fatalf("timestamp at index %d: expected %v, got %v", i, index[i], dp.Timestamp)
		}
		if dp.Value != 3.5 {
			t.Fatalf("value at index %d: expected %f, got %f", i, 3.5, dp.Value)
		}
	}
}

func TestUniformNoiseBounds(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := MakeSeriesIndex(base, time.Second, 1000)

	ts := UniformNoise(index, -2, 4)
	for i, dp := range ts.DataPoints() {
		if !dp.Timestamp.Equal(index[i]) {
			t.Fatalf("timestamp at index %d: expected %v, got %v", i, index[i], dp.Timestamp)
		}
		if dp.Value < -2 || dp.Value > 4 {
			t.Fatalf("value at index %d out of bounds: %f", i, dp.Value)
		}
	}
}
