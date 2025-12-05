package timeseriesgo

import (
	"testing"
	"time"
)

func TestZScore(t *testing.T) {
	ts := Empty()
	now := time.Now()
	values := []float64{10, 11, 10, 12, 11, 50}
	for i, v := range values {
		ts.AddPoint(DataPoint{timestamp: now.Add(time.Duration(i) * time.Hour), value: v})
	}
	zscored, err := ZScore(ts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedValues := []float64{-0.457738, -0.395319, -0.457738, -0.332900, -0.395319, 2.039013}

	if zscored.Length() != ts.Length() {
		t.Errorf("Expected Z-Scored series length %d, got %d", ts.Length(), zscored.Length())
	}

	for i, dp := range zscored.datapoints {
		if dp.value-expectedValues[i] > 0.0001 || expectedValues[i]-dp.value > 0.0001 {
			t.Errorf("At index %d: expected Z-Score value %f, got %f", i, expectedValues[i], dp.value)
		}
	}

	anomaly, err2 := FindAnomaliesWithZScore(ts)
	if err2 != nil {
		t.Errorf("Unexpected error: %v", err2)
	}
	expectedAnomalies := []float64{0, 0, 0, 0, 0, 1}

	for i, dp := range anomaly.datapoints {
		if dp.value != expectedAnomalies[i] {
			t.Errorf("At index %d: expected Anomaly value %f, got %f", i, expectedAnomalies[i], dp.value)
		}
	}
}

func TestRobustZScore(t *testing.T) {
	ts := Empty()
	now := time.Now()
	values := []float64{10, 11, 10, 12, 11, 50}
	for i, v := range values {
		ts.AddPoint(DataPoint{timestamp: now.Add(time.Duration(i) * time.Hour), value: v})
	}
	zscored, err := RobustZScore(ts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedValues := []float64{-0.674491, 0.0, -0.674491, 0.674491, 0.0, 26.305140}

	if zscored.Length() != ts.Length() {
		t.Errorf("Expected Robust Z-Scored series length %d, got %d", ts.Length(), zscored.Length())
	}

	for i, dp := range zscored.datapoints {
		if dp.value-expectedValues[i] > 0.0001 || expectedValues[i]-dp.value > 0.0001 {
			t.Errorf("At index %d: expected Robust Z-Score value %f, got %f", i, expectedValues[i], dp.value)
		}
	}

	anomaly, err2 := FindAnomaliesWithRobustZScore(ts)
	if err2 != nil {
		t.Errorf("Unexpected error: %v", err2)
	}
	expectedAnomalies := []float64{0, 0, 0, 0, 0, 1}

	for i, dp := range anomaly.datapoints {
		if dp.value != expectedAnomalies[i] {
			t.Errorf("At index %d: expected Anomaly value %f, got %f", i, expectedAnomalies[i], dp.value)
		}
	}
}
