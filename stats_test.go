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
	expectedSampleVariance := 17.2
	expectedPopulationVariance := 14.333333

	if mv.mean != expectedMean {
		t.Errorf("Expected mean %f, got %f", expectedMean, mv.mean)
	}
	if mv.sampleVariance-expectedSampleVariance > 0.0001 {
		t.Errorf("Expected sample variance %f, got %f", expectedSampleVariance, mv.sampleVariance)
	}
	if mv.populationVariance-expectedPopulationVariance > 0.0001 {
		t.Errorf("Expected population variance %f, got %f", expectedPopulationVariance, mv.populationVariance)
	}
}

func TestMovingAverageWindow(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	ts := Empty()
	ts.AddPoint(DataPoint{timestamp: base, value: 1})
	ts.AddPoint(DataPoint{timestamp: base.Add(time.Minute), value: 3})
	ts.AddPoint(DataPoint{timestamp: base.Add(2 * time.Minute), value: 5})

	window := 2 * time.Minute
	ma := ts.MovingAverage(window)

	if ma.Length() != ts.Length() {
		t.Fatalf("Expected moving average length %d, got %d", ts.Length(), ma.Length())
	}

	expected := []float64{1, 2, 4}

	for i, dp := range ma.datapoints {
		if dp.timestamp != ts.datapoints[i].timestamp {
			t.Errorf("At idx %d expected timestamp %v, got %v", i, ts.datapoints[i].timestamp, dp.timestamp)
		}
		if dp.value != expected[i] {
			t.Errorf("At idx %d expected value %.1f, got %.1f", i, expected[i], dp.value)
		}
	}
}
