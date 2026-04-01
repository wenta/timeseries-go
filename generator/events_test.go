package generator

import (
	"testing"
	"time"
)

func TestPulseTrain(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := MakeSeriesIndex(base, time.Minute, 8)

	ts := PulseTrain(index, []Pulse{
		{StartIndex: 1, Duration: 3, Amplitude: 2},
		{StartIndex: 2, Duration: 2, Amplitude: 1},
		{StartIndex: -2, Duration: 3, Amplitude: 5},
		{StartIndex: 6, Duration: 5, Amplitude: 4},
		{StartIndex: 4, Duration: 0, Amplitude: 9},
	})

	expected := []float64{5, 2, 3, 3, 0, 0, 4, 4}
	for i, dp := range ts.DataPoints() {
		if !dp.Timestamp.Equal(index[i]) {
			t.Fatalf("timestamp at index %d: expected %v, got %v", i, index[i], dp.Timestamp)
		}
		if dp.Value != expected[i] {
			t.Fatalf("value at index %d: expected %f, got %f", i, expected[i], dp.Value)
		}
	}
}

func TestPulseTrainEmptyIndex(t *testing.T) {
	if result := PulseTrain(nil, []Pulse{{StartIndex: 0, Duration: 2, Amplitude: 1}}); !result.IsEmpty() {
		t.Fatalf("expected empty series, got length %d", result.Length())
	}
}

func TestPoissonEventIndices(t *testing.T) {
	indices := PoissonEventIndices(100, 0.7)
	for i, index := range indices {
		if index < 0 || index >= 100 {
			t.Fatalf("index out of range: %d", index)
		}
		if i > 0 && indices[i-1] >= index {
			t.Fatalf("indices not strictly ascending at position %d: %v", i, indices)
		}
	}
}

func TestPoissonEventIndicesEdgeCases(t *testing.T) {
	if indices := PoissonEventIndices(0, 0.5); len(indices) != 0 {
		t.Fatalf("expected no events for zero length, got %v", indices)
	}
	if indices := PoissonEventIndices(10, 0); len(indices) != 0 {
		t.Fatalf("expected no events for zero lambda, got %v", indices)
	}
	if indices := PoissonEventIndices(10, -1); len(indices) != 0 {
		t.Fatalf("expected no events for negative lambda, got %v", indices)
	}
}
