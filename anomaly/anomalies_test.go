package anomaly

import (
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

func TestZScore(t *testing.T) {
	ts := timeseriesgo.Empty()
	now := time.Now()
	values := []float64{10, 11, 10, 12, 11, 50}
	for i, v := range values {
		ts.AddPoint(timeseriesgo.DataPoint{Timestamp: now.Add(time.Duration(i) * time.Hour), Value: v})
	}
	zscored, err := ZScore(ts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedValues := []float64{-0.457738, -0.395319, -0.457738, -0.332900, -0.395319, 2.039013}

	if zscored.Length() != ts.Length() {
		t.Errorf("Expected Z-Scored series length %d, got %d", ts.Length(), zscored.Length())
	}

	for i, dp := range zscored.DataPoints() {
		if dp.Value-expectedValues[i] > 0.0001 || expectedValues[i]-dp.Value > 0.0001 {
			t.Errorf("At index %d: expected Z-Score value %f, got %f", i, expectedValues[i], dp.Value)
		}
	}

	anomaly, err2 := FindAnomaliesWithZScore(ts)
	if err2 != nil {
		t.Errorf("Unexpected error: %v", err2)
	}
	expectedAnomalies := []float64{0, 0, 0, 0, 0, 1}

	for i, dp := range anomaly.DataPoints() {
		if dp.Value != expectedAnomalies[i] {
			t.Errorf("At index %d: expected Anomaly value %f, got %f", i, expectedAnomalies[i], dp.Value)
		}
	}
}

func TestRobustZScore(t *testing.T) {
	ts := timeseriesgo.Empty()
	now := time.Now()
	values := []float64{10, 11, 10, 12, 11, 50}
	for i, v := range values {
		ts.AddPoint(timeseriesgo.DataPoint{Timestamp: now.Add(time.Duration(i) * time.Hour), Value: v})
	}
	zscored, err := RobustZScore(ts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedValues := []float64{-0.674491, 0.0, -0.674491, 0.674491, 0.0, 26.305140}

	if zscored.Length() != ts.Length() {
		t.Errorf("Expected Robust Z-Scored series length %d, got %d", ts.Length(), zscored.Length())
	}

	for i, dp := range zscored.DataPoints() {
		if dp.Value-expectedValues[i] > 0.0001 || expectedValues[i]-dp.Value > 0.0001 {
			t.Errorf("At index %d: expected Robust Z-Score value %f, got %f", i, expectedValues[i], dp.Value)
		}
	}

	anomaly, err2 := FindAnomaliesWithRobustZScore(ts)
	if err2 != nil {
		t.Errorf("Unexpected error: %v", err2)
	}
	expectedAnomalies := []float64{0, 0, 0, 0, 0, 1}

	for i, dp := range anomaly.DataPoints() {
		if dp.Value != expectedAnomalies[i] {
			t.Errorf("At index %d: expected Anomaly value %f, got %f", i, expectedAnomalies[i], dp.Value)
		}
	}
}

func TestFindSpikeAnomalies(t *testing.T) {
	ts := timeseriesgo.Empty()
	now := time.Now()
	values := []float64{10, 11, 30, 29, 29, 29, 20, 21}
	for i, v := range values {
		ts.AddPoint(timeseriesgo.DataPoint{Timestamp: now.Add(time.Duration(i) * time.Minute), Value: v})
	}

	anomalies, err := FindSpikeAnomalies(ts, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []float64{0, 0, 1, 0, 0, 0, 0, 0}
	for i, dp := range anomalies.DataPoints() {
		if dp.Value != expected[i] {
			t.Errorf("At index %d: expected Anomaly value %f, got %f", i, expected[i], dp.Value)
		}
	}
}

func TestFindDropAnomalies(t *testing.T) {
	ts := timeseriesgo.Empty()
	now := time.Now()
	values := []float64{10, 11, 30, 29, 29, 29, 20, 21}
	for i, v := range values {
		ts.AddPoint(timeseriesgo.DataPoint{Timestamp: now.Add(time.Duration(i) * time.Minute), Value: v})
	}

	anomalies, err := FindDropAnomalies(ts, 8)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []float64{0, 0, 0, 0, 0, 0, 1, 0}
	for i, dp := range anomalies.DataPoints() {
		if dp.Value != expected[i] {
			t.Errorf("At index %d: expected Anomaly value %f, got %f", i, expected[i], dp.Value)
		}
	}
}

func TestFindFlatlineAnomalies(t *testing.T) {
	ts := timeseriesgo.Empty()
	now := time.Now()
	values := []float64{10, 11, 30, 29, 29, 29, 20, 21}
	for i, v := range values {
		ts.AddPoint(timeseriesgo.DataPoint{Timestamp: now.Add(time.Duration(i) * time.Minute), Value: v})
	}

	anomalies, err := FindFlatlineAnomalies(ts, 0.01, 3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []float64{0, 0, 0, 1, 1, 1, 0, 0}
	for i, dp := range anomalies.DataPoints() {
		if dp.Value != expected[i] {
			t.Errorf("At index %d: expected Anomaly value %f, got %f", i, expected[i], dp.Value)
		}
	}
}
