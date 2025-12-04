package timeseriesgo

import (
	"testing"
	"time"
)

func TestMeanAndVariance(t *testing.T) {
	ts := Empty()

	ts.AddPoint(DataPoint{timestamp: time.Now(), value: 1.0})
	ts.AddPoint(DataPoint{timestamp: time.Now().Add(time.Minute), value: -3.0})
	ts.AddPoint(DataPoint{timestamp: time.Now().Add(2 * time.Minute), value: 6.0})
	ts.AddPoint(DataPoint{timestamp: time.Now().Add(3 * time.Minute), value: 6.0})
	ts.AddPoint(DataPoint{timestamp: time.Now().Add(4 * time.Minute), value: 6.0})
	ts.AddPoint(DataPoint{timestamp: time.Now().Add(5 * time.Minute), value: 8.0})

	mv, err := ts.GetMeanAndVariance()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedMean := 4.0
	expectedVariance := 14.333333

	if mv.mean != expectedMean {
		t.Errorf("Expected mean %f, got %f", expectedMean, mv.mean)
	}
	if mv.variance-expectedVariance > 0.0001 {
		t.Errorf("Expected variance %f, got %f", expectedVariance, mv.variance)
	}
}
