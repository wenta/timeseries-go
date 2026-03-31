package decompose

import (
	"math"
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/generator"
)

func TestSTLAdditiveSeries(t *testing.T) {
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	period := 12
	seasonalPattern := []float64{-6, -4, -2, 0, 2, 4, 6, 4, 2, 0, -2, -4}
	values := make([]float64, 60)
	for i := range values {
		trend := 100.0 + 1.5*float64(i)
		values[i] = trend + seasonalPattern[i%period]
	}

	index := make([]time.Time, len(values))
	for i := range index {
		index[i] = base.AddDate(0, i, 0)
	}

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("unexpected error creating series: %v", err)
	}

	result, err := STL(ts, STLConfig{
		Period:          period,
		Seasonal:        7,
		Trend:           21,
		LowPass:         13,
		InnerIterations: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Trend.Length() != ts.Length() || result.Seasonal.Length() != ts.Length() || result.Residual.Length() != ts.Length() {
		t.Fatalf("expected all components to match input length %d", ts.Length())
	}

	const tolerance = 1.5
	trendPoints := result.Trend.DataPoints()
	seasonalPoints := result.Seasonal.DataPoints()
	residualPoints := result.Residual.DataPoints()
	for i := period; i < len(values)-period; i++ {
		expectedTrend := 100.0 + 1.5*float64(i)
		if math.Abs(trendPoints[i].Value-expectedTrend) > tolerance {
			t.Fatalf("trend at index %d: expected around %f, got %f", i, expectedTrend, trendPoints[i].Value)
		}
		if math.Abs(seasonalPoints[i].Value-seasonalPattern[i%period]) > tolerance {
			t.Fatalf("seasonal at index %d: expected around %f, got %f", i, seasonalPattern[i%period], seasonalPoints[i].Value)
		}
		if math.Abs(residualPoints[i].Value) > tolerance {
			t.Fatalf("residual at index %d: expected around 0, got %f", i, residualPoints[i].Value)
		}
	}
}

func TestSTLRobustDownweightsOutlier(t *testing.T) {
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	period := 12
	seasonalPattern := []float64{-6, -4, -2, 0, 2, 4, 6, 4, 2, 0, -2, -4}
	values := make([]float64, 60)
	for i := range values {
		trend := 100.0 + 1.5*float64(i)
		values[i] = trend + seasonalPattern[i%period]
	}
	values[30] += 40

	index := make([]time.Time, len(values))
	for i := range index {
		index[i] = base.AddDate(0, i, 0)
	}

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("unexpected error creating series: %v", err)
	}

	plain, err := STL(ts, STLConfig{
		Period:          period,
		Seasonal:        7,
		Trend:           21,
		LowPass:         13,
		InnerIterations: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	robust, err := STL(ts, STLConfig{
		Period:          period,
		Seasonal:        7,
		Trend:           21,
		LowPass:         13,
		Robust:          true,
		InnerIterations: 2,
		OuterIterations: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedTrend := 100.0 + 1.5*30
	plainDistance := math.Abs(plain.Trend.DataPoints()[30].Value - expectedTrend)
	robustDistance := math.Abs(robust.Trend.DataPoints()[30].Value - expectedTrend)
	if robustDistance >= plainDistance {
		t.Fatalf("expected robust trend to be closer to baseline at outlier point, plain=%f robust=%f", plainDistance, robustDistance)
	}

}

func TestSTLDefaultsAndInvalidInput(t *testing.T) {
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	index := generator.MakeSeriesIndex(base, time.Hour, 24)
	values := make([]float64, len(index))
	for i := range values {
		values[i] = float64(i)
	}
	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("unexpected error creating series: %v", err)
	}

	if _, err := STL(ts, STLConfig{Period: 6}); err != nil {
		t.Fatalf("expected default config to work, got %v", err)
	}

	cases := []struct {
		name   string
		ts     timeseriesgo.TimeSeries
		config STLConfig
	}{
		{name: "empty", ts: timeseriesgo.Empty(), config: STLConfig{Period: 6}},
		{name: "period too small", ts: ts, config: STLConfig{Period: 1}},
		{name: "not enough cycles", ts: ts, config: STLConfig{Period: 20}},
		{name: "even seasonal", ts: ts, config: STLConfig{Period: 6, Seasonal: 8}},
		{name: "even trend", ts: ts, config: STLConfig{Period: 6, Trend: 10}},
		{name: "even lowpass", ts: ts, config: STLConfig{Period: 6, LowPass: 8}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := STL(tc.ts, tc.config); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
