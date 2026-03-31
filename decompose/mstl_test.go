package decompose

import (
	"math"
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/generator"
)

func TestMSTLTwoSeasonalities(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dailyPeriod := 24
	weeklyPeriod := 24 * 7

	length := weeklyPeriod * 3
	dailyPattern := make([]float64, dailyPeriod)
	for i := range dailyPattern {
		dailyPattern[i] = 5 * math.Sin(2*math.Pi*float64(i)/float64(dailyPeriod))
	}
	weeklyPattern := make([]float64, weeklyPeriod)
	for i := range weeklyPattern {
		day := (i / dailyPeriod) % 7
		weeklyPattern[i] = []float64{-4, -2, 0, 1, 3, 5, -3}[day]
	}

	values := make([]float64, length)
	for i := range values {
		trend := 50.0 + 0.05*float64(i)
		values[i] = trend + dailyPattern[i%dailyPeriod] + weeklyPattern[i%weeklyPeriod]
	}

	index := generator.MakeSeriesIndex(base, time.Hour, length)
	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("unexpected error creating series: %v", err)
	}

	result, err := MSTL(ts, MSTLConfig{
		Periods:         []int{dailyPeriod, weeklyPeriod},
		SeasonalWindows: []int{9, 15},
		Iterations:      2,
		InnerIterations: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Seasonal) != 2 {
		t.Fatalf("expected 2 seasonal components, got %d", len(result.Seasonal))
	}

	daily := result.Seasonal[0]
	weekly := result.Seasonal[1]
	if daily.Period != dailyPeriod || weekly.Period != weeklyPeriod {
		t.Fatalf("unexpected component periods: got %d and %d", daily.Period, weekly.Period)
	}

	const tolerance = 2.0
	for i := weeklyPeriod; i < length-weeklyPeriod; i++ {
		expectedTrend := 50.0 + 0.05*float64(i)
		if math.Abs(result.Trend.DataPoints()[i].Value-expectedTrend) > tolerance {
			t.Fatalf("trend at index %d: expected around %f, got %f", i, expectedTrend, result.Trend.DataPoints()[i].Value)
		}
		if math.Abs(daily.Series.DataPoints()[i].Value-dailyPattern[i%dailyPeriod]) > tolerance {
			t.Fatalf("daily component at index %d: expected around %f, got %f", i, dailyPattern[i%dailyPeriod], daily.Series.DataPoints()[i].Value)
		}
		if math.Abs(weekly.Series.DataPoints()[i].Value-weeklyPattern[i%weeklyPeriod]) > tolerance {
			t.Fatalf("weekly component at index %d: expected around %f, got %f", i, weeklyPattern[i%weeklyPeriod], weekly.Series.DataPoints()[i].Value)
		}
		if math.Abs(result.Residual.DataPoints()[i].Value) > tolerance {
			t.Fatalf("residual at index %d: expected around 0, got %f", i, result.Residual.DataPoints()[i].Value)
		}
	}
}

func TestMSTLInvalidInput(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	index := generator.MakeSeriesIndex(base, time.Hour, 48)
	values := make([]float64, len(index))
	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("unexpected error creating series: %v", err)
	}

	cases := []struct {
		name   string
		ts     timeseriesgo.TimeSeries
		config MSTLConfig
	}{
		{name: "empty", ts: timeseriesgo.Empty(), config: MSTLConfig{Periods: []int{24}}},
		{name: "no periods", ts: ts, config: MSTLConfig{}},
		{name: "period too small", ts: ts, config: MSTLConfig{Periods: []int{1}}},
		{name: "duplicate periods", ts: ts, config: MSTLConfig{Periods: []int{12, 12}}},
		{name: "not enough data", ts: ts, config: MSTLConfig{Periods: []int{30}}},
		{name: "window length mismatch", ts: ts, config: MSTLConfig{Periods: []int{12, 24}, SeasonalWindows: []int{9}}},
		{name: "even seasonal window", ts: ts, config: MSTLConfig{Periods: []int{12}, SeasonalWindows: []int{8}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := MSTL(tc.ts, tc.config); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
