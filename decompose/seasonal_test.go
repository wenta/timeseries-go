package decompose

import (
	"math"
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/generator"
)

func TestSeasonalDecomposeAdditiveEvenPeriod(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	period := 4
	seasonalPattern := []float64{-1, 2, -2, 1}
	values := make([]float64, 16)
	for i := range values {
		trend := 10.0 + 2.0*float64(i)
		values[i] = trend + seasonalPattern[i%period]
	}

	index := generator.MakeSeriesIndex(base, time.Hour, len(values))
	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("unexpected error creating series: %v", err)
	}

	result, err := SeasonalDecompose(ts, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSameTimestamps(t, ts, result.Trend)
	assertSameTimestamps(t, ts, result.Seasonal)
	assertSameTimestamps(t, ts, result.Residual)

	const epsilon = 1e-9
	for i := range values {
		expectedTrend := 10.0 + 2.0*float64(i)
		if math.Abs(result.Trend.DataPoints()[i].Value-expectedTrend) > epsilon {
			t.Fatalf("trend at index %d: expected %f, got %f", i, expectedTrend, result.Trend.DataPoints()[i].Value)
		}
		if math.Abs(result.Seasonal.DataPoints()[i].Value-seasonalPattern[i%period]) > epsilon {
			t.Fatalf("seasonal at index %d: expected %f, got %f", i, seasonalPattern[i%period], result.Seasonal.DataPoints()[i].Value)
		}
		if math.Abs(result.Residual.DataPoints()[i].Value) > epsilon {
			t.Fatalf("residual at index %d: expected 0, got %f", i, result.Residual.DataPoints()[i].Value)
		}
	}
}

func TestSeasonalDecomposeAdditiveOddPeriod(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	period := 3
	seasonalPattern := []float64{-1, 0, 1}
	values := make([]float64, 12)
	for i := range values {
		trend := 5.0 + float64(i)
		values[i] = trend + seasonalPattern[i%period]
	}

	index := generator.MakeSeriesIndex(base, time.Hour, len(values))
	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("unexpected error creating series: %v", err)
	}

	result, err := SeasonalDecompose(ts, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const epsilon = 1e-9
	for i := range values {
		expectedTrend := 5.0 + float64(i)
		if math.Abs(result.Trend.DataPoints()[i].Value-expectedTrend) > epsilon {
			t.Fatalf("trend at index %d: expected %f, got %f", i, expectedTrend, result.Trend.DataPoints()[i].Value)
		}
		if math.Abs(result.Seasonal.DataPoints()[i].Value-seasonalPattern[i%period]) > epsilon {
			t.Fatalf("seasonal at index %d: expected %f, got %f", i, seasonalPattern[i%period], result.Seasonal.DataPoints()[i].Value)
		}
		if math.Abs(result.Residual.DataPoints()[i].Value) > epsilon {
			t.Fatalf("residual at index %d: expected 0, got %f", i, result.Residual.DataPoints()[i].Value)
		}
	}
}

func TestSeasonalDecomposeInvalidInput(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	index := generator.MakeSeriesIndex(base, time.Hour, 6)
	values := []float64{1, 2, 3, 4, 5, 6}
	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("unexpected error creating series: %v", err)
	}

	cases := []struct {
		name   string
		ts     timeseriesgo.TimeSeries
		period int
	}{
		{name: "empty", ts: timeseriesgo.Empty(), period: 4},
		{name: "period too small", ts: ts, period: 1},
		{name: "not enough cycles", ts: ts, period: 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := SeasonalDecompose(tc.ts, tc.period); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func assertSameTimestamps(t *testing.T, expected timeseriesgo.TimeSeries, actual timeseriesgo.TimeSeries) {
	t.Helper()

	if expected.Length() != actual.Length() {
		t.Fatalf("expected series length %d, got %d", expected.Length(), actual.Length())
	}

	expectedPoints := expected.DataPoints()
	actualPoints := actual.DataPoints()

	for i := range expectedPoints {
		if !expectedPoints[i].Timestamp.Equal(actualPoints[i].Timestamp) {
			t.Fatalf("timestamp at index %d: expected %v, got %v", i, expectedPoints[i].Timestamp, actualPoints[i].Timestamp)
		}
	}
}
