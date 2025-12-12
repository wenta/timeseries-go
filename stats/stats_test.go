package stats

import (
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

func TestMeanAndVariance(t *testing.T) {
	ts := timeseriesgo.Empty()

	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Now(), Value: 1.0})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Now().Add(time.Minute), Value: -3.0})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Now().Add(2 * time.Minute), Value: 6.0})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Now().Add(3 * time.Minute), Value: 6.0})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Now().Add(4 * time.Minute), Value: 6.0})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Now().Add(5 * time.Minute), Value: 8.0})

	mv, err := GetMeanAndVariance(ts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedMean := 4.0
	expectedSampleVariance := 17.2
	expectedPopulationVariance := 14.333333

	if mv.Mean != expectedMean {
		t.Errorf("Expected mean %f, got %f", expectedMean, mv.Mean)
	}
	if mv.SampleVariance-expectedSampleVariance > 0.0001 {
		t.Errorf("Expected sample variance %f, got %f", expectedSampleVariance, mv.SampleVariance)
	}
	if mv.PopulationVariance-expectedPopulationVariance > 0.0001 {
		t.Errorf("Expected population variance %f, got %f", expectedPopulationVariance, mv.PopulationVariance)
	}
}

func TestMovingAverageWindow(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	ts := timeseriesgo.Empty()
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 1})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(time.Minute), Value: 3})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(2 * time.Minute), Value: 5})

	window := 2 * time.Minute
	ma := MovingAverage(ts, window)

	if ma.Length() != ts.Length() {
		t.Fatalf("Expected moving average length %d, got %d", ts.Length(), ma.Length())
	}

	expected := []float64{1, 2, 4}

	for i, dp := range ma.DataPoints() {
		tsPoints := ts.DataPoints()
		if dp.Timestamp != tsPoints[i].Timestamp {
			t.Errorf("At idx %d expected timestamp %v, got %v", i, tsPoints[i].Timestamp, dp.Timestamp)
		}
		if dp.Value != expected[i] {
			t.Errorf("At idx %d expected value %.1f, got %.1f", i, expected[i], dp.Value)
		}
	}
}
