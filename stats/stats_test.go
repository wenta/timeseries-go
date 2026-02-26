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

func TestCorrelation(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	
	ts1 := timeseriesgo.Empty()
	ts2 := timeseriesgo.Empty()

	values1 := []float64{1, 2, 3}
	values2 := []float64{2, 4, 6}

	for i := 0; i < 3; i++ {
		ts1.AddPoint(timeseriesgo.DataPoint{
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Value:     values1[i],
		})
		ts2.AddPoint(timeseriesgo.DataPoint{
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Value:     values2[i],
		})
	}

	corr, err := Correlation(ts1, ts2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if corr < 0.999 {
		t.Errorf("expected correlation close to 1, got %f", corr)
	}
}

func TestCorrelationEdgeCases(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	
	ts1 := timeseriesgo.Empty()
	ts2 := timeseriesgo.Empty()

	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 5})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(time.Minute), Value: 5})

	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 1})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(time.Minute), Value: 2})

	if _, err := Correlation(ts1, ts2); err == nil {
		t.Errorf("expected error for zero variance")
	}

	
	ts3 := timeseriesgo.Empty()
	ts4 := timeseriesgo.Empty()

	ts3.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 1})
	ts4.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 2})

	if _, err := Correlation(ts3, ts4); err == nil {
		t.Errorf("expected error for insufficient points")
	}
}