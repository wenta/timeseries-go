package stats

import (
	"math"
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

func TestACF(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	ts := timeseriesgo.Empty()
	values := []float64{1, 2, 3, 4}
	for i, value := range values {
		ts.AddPoint(timeseriesgo.DataPoint{
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Value:     value,
		})
	}

	acf, err := ACF(ts, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []float64{1.0, 0.25, -0.3, -0.45}
	const epsilon = 1e-9
	for i := range expected {
		if math.Abs(acf[i]-expected[i]) > epsilon {
			t.Fatalf("lag %d: expected %f, got %f", i, expected[i], acf[i])
		}
	}
}

func TestACFInvalidInput(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	onePoint := timeseriesgo.Empty()
	onePoint.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 1})

	constant := timeseriesgo.Empty()
	constant.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 5})
	constant.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(time.Minute), Value: 5})
	constant.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(2 * time.Minute), Value: 5})

	valid := timeseriesgo.Empty()
	valid.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 1})
	valid.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(time.Minute), Value: 2})
	valid.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(2 * time.Minute), Value: 3})

	cases := []struct {
		name  string
		ts    timeseriesgo.TimeSeries
		nlags int
	}{
		{name: "empty", ts: timeseriesgo.Empty(), nlags: 1},
		{name: "negative lags", ts: valid, nlags: -1},
		{name: "single point", ts: onePoint, nlags: 0},
		{name: "too many lags", ts: valid, nlags: 3},
		{name: "zero variance", ts: constant, nlags: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ACF(tc.ts, tc.nlags); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestMinMaxNormalize(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	ts := timeseriesgo.Empty()
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 0})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(time.Minute), Value: 5})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(2 * time.Minute), Value: 10})

	result, err := MinMaxNormalize(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []float64{0.0, 0.5, 1.0}
	for i, dp := range result.DataPoints() {
		tsPoints := ts.DataPoints()
		if dp.Timestamp != tsPoints[i].Timestamp {
			t.Errorf("at idx %d: expected timestamp %v, got %v", i, tsPoints[i].Timestamp, dp.Timestamp)
		}
		if dp.Value-expected[i] > 0.0001 {
			t.Errorf("at idx %d: expected value %.4f, got %.4f", i, expected[i], dp.Value)
		}
	}
}

func TestMinMaxNormalizeAllEqual(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	ts := timeseriesgo.Empty()
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 7})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(time.Minute), Value: 7})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(2 * time.Minute), Value: 7})

	result, err := MinMaxNormalize(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, dp := range result.DataPoints() {
		if dp.Value != 0.0 {
			t.Errorf("at idx %d: expected 0.0 for equal values, got %f", i, dp.Value)
		}
	}
}

func TestMinMaxNormalizeEmpty(t *testing.T) {
	ts := timeseriesgo.Empty()
	_, err := MinMaxNormalize(ts)
	if err == nil {
		t.Error("expected error for empty TimeSeries, got nil")
	}
}

func TestMinMaxNormalizeSinglePoint(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	ts := timeseriesgo.Empty()
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 42})

	result, err := MinMaxNormalize(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	points := result.DataPoints()
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
	if points[0].Value != 0.0 {
		t.Errorf("expected 0.0 for single point, got %f", points[0].Value)
	}
}

func TestMinMaxNormalizeNegativeValues(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	ts := timeseriesgo.Empty()
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: -10})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(time.Minute), Value: 0})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(2 * time.Minute), Value: 10})

	result, err := MinMaxNormalize(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []float64{0.0, 0.5, 1.0}
	for i, dp := range result.DataPoints() {
		if dp.Value-expected[i] > 0.0001 {
			t.Errorf("at idx %d: expected %.4f, got %.4f", i, expected[i], dp.Value)
		}
	}
}
