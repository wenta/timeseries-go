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
