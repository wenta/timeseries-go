package timeseriesgo

import (
	"testing"
	"time"
)

func TestRandomNoiseLength(t *testing.T) {
	start := time.Now()
	step := time.Minute

	ts := RandomNoise(start, 10, step)

	if ts.Length() != 10 {
		t.Errorf("expected 10 points, got %d", ts.Length())
	}
}

func TestRandomNoiseStep(t *testing.T) {
	start := time.Now()
	step := time.Minute

	ts := RandomNoise(start, 5, step)

	points := ts.DataPoints()

	for i := 1; i < len(points); i++ {
		diff := points[i].Timestamp.Sub(points[i-1].Timestamp)
		if diff != step {
			t.Errorf("expected step %v, got %v", step, diff)
		}
	}
}

func TestRandomNoiseZero(t *testing.T) {
	start := time.Now()
	step := time.Minute

	ts := RandomNoise(start, 0, step)

	if !ts.IsEmpty() {
		t.Errorf("expected empty time series when n <= 0")
	}
}
