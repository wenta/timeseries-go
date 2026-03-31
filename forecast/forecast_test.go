package forecast

import (
	"math"
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/generator"
)

func TestNaiveForecast(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Hour
	count := 5
	index := generator.MakeSeriesIndex(start, interval, count)

	ts := generator.RandomWalk(index, 5)
	forecastHorizon := 3
	forecast := Naive(ts, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Errorf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	lastPoint, err := ts.Last()
	if err != nil {
		t.Errorf("Unexpected error getting last point: %v", err)
	}

	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if dp.Value != lastPoint.Value {
			t.Errorf("At index %d: expected value %f, got %f", i, lastPoint.Value, dp.Value)
		}
	}
}

func TestSimpleExponentialSmoothing(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Hour
	values := []float64{10, 12, 13, 12}
	index := generator.MakeSeriesIndex(start, interval, len(values))

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	forecastHorizon := 3
	alpha := 0.5
	forecast := SimpleExponentialSmoothing(ts, alpha, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Errorf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	expectedValue := 12.0
	lastPoint, err := ts.Last()
	if err != nil {
		t.Fatalf("Unexpected error getting last point: %v", err)
	}
	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if dp.Value != expectedValue {
			t.Errorf("At index %d: expected value %f, got %f", i, expectedValue, dp.Value)
		}
	}
}

func TestSimpleExponentialSmoothingStatsmodelsOilData(t *testing.T) {
	// Data from the statsmodels exponential smoothing example (Saudi Arabia oil).
	values := []float64{
		446.6565,
		454.4733,
		455.663,
		423.6322,
		456.2713,
		440.5881,
		425.3325,
		485.1494,
		506.0482,
		526.792,
		514.2689,
		494.211,
	}
	start := time.Date(1996, 12, 31, 0, 0, 0, 0, time.UTC)
	interval := 365 * 24 * time.Hour
	index := generator.MakeSeriesIndex(start, interval, len(values))

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	forecastHorizon := 3
	alpha := 0.2
	forecast := SimpleExponentialSmoothing(ts, alpha, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Errorf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	expectedValue := 484.80246538161776
	const epsilon = 1e-9
	lastPoint, err := ts.Last()
	if err != nil {
		t.Fatalf("Unexpected error getting last point: %v", err)
	}

	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if math.Abs(dp.Value-expectedValue) > epsilon {
			t.Errorf("At index %d: expected value %f, got %f", i, expectedValue, dp.Value)
		}
	}
}

func TestDoubleExponentialSmoothingLinearTrend(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Hour
	values := []float64{10, 12, 14, 16}
	index := generator.MakeSeriesIndex(start, interval, len(values))

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	forecastHorizon := 3
	forecast := DoubleExponentialSmoothing(ts, 1.0, 1.0, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Fatalf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	expectedValues := []float64{18, 20, 22}
	lastPoint, err := ts.Last()
	if err != nil {
		t.Fatalf("Unexpected error getting last point: %v", err)
	}

	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if dp.Value != expectedValues[i] {
			t.Errorf("At index %d: expected value %f, got %f", i, expectedValues[i], dp.Value)
		}
	}
}

func TestDoubleExponentialSmoothingHandCalculated(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Hour
	values := []float64{10, 12, 13, 16}
	index := generator.MakeSeriesIndex(start, interval, len(values))

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	forecastHorizon := 3
	alpha := 0.5
	beta := 0.4
	forecast := DoubleExponentialSmoothing(ts, alpha, beta, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Fatalf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	expectedValues := []float64{17.59, 19.53, 21.47}
	const epsilon = 1e-9
	lastPoint, err := ts.Last()
	if err != nil {
		t.Fatalf("Unexpected error getting last point: %v", err)
	}

	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if math.Abs(dp.Value-expectedValues[i]) > epsilon {
			t.Errorf("At index %d: expected value %f, got %f", i, expectedValues[i], dp.Value)
		}
	}
}

func TestDoubleExponentialSmoothingLiteratureAirDataSimpleInitialization(t *testing.T) {
	// Air pollution series from the statsmodels exponential smoothing example,
	// which reproduces examples from Hyndman and Athanasopoulos,
	// Forecasting: Principles and Practice.
	// Expected values below come from the Holt recursions with this package's
	// simple initialization: level=y1, trend=y2-y1.
	values := []float64{
		17.5534,
		21.86,
		23.8866,
		26.9293,
		26.8885,
		28.8314,
		30.0751,
		30.9535,
		30.1857,
		31.5797,
		32.5776,
		33.4774,
		39.0216,
		41.3864,
		41.5966,
	}
	start := time.Date(1990, 12, 31, 0, 0, 0, 0, time.UTC)
	interval := 365 * 24 * time.Hour
	index := generator.MakeSeriesIndex(start, interval, len(values))

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	forecastHorizon := 5
	alpha := 0.8
	beta := 0.2
	forecast := DoubleExponentialSmoothing(ts, alpha, beta, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Fatalf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	expectedValues := []float64{
		43.80344184942457,
		45.67461174639456,
		47.54578164336455,
		49.41695154033453,
		51.28812143730452,
	}
	const epsilon = 1e-9
	lastPoint, err := ts.Last()
	if err != nil {
		t.Fatalf("Unexpected error getting last point: %v", err)
	}

	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if math.Abs(dp.Value-expectedValues[i]) > epsilon {
			t.Errorf("At index %d: expected value %f, got %f", i, expectedValues[i], dp.Value)
		}
	}
}

func TestDoubleExponentialSmoothingEstimatedShortSeriesUsesSimpleEstimatedInitialization(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Hour
	values := []float64{10, 12, 13, 16}
	index := generator.MakeSeriesIndex(start, interval, len(values))

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	alpha := 0.5
	beta := 0.4
	forecastHorizon := 3

	forecast := DoubleExponentialSmoothingEstimated(ts, alpha, beta, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Fatalf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	expectedValues := []float64{16.9382, 18.5194, 20.1006}
	const epsilon = 1e-12
	lastPoint, err := ts.Last()
	if err != nil {
		t.Fatalf("Unexpected error getting last point: %v", err)
	}

	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if math.Abs(dp.Value-expectedValues[i]) > epsilon {
			t.Errorf("At index %d: expected value %f, got %f", i, expectedValues[i], dp.Value)
		}
	}
}

func TestDoubleExponentialSmoothingEstimatedLiteratureAirData(t *testing.T) {
	// Air pollution series from the statsmodels exponential smoothing example,
	// which reproduces examples from Hyndman and Athanasopoulos,
	// Forecasting: Principles and Practice.
	// Expected values were reproduced against:
	// Holt(values, initialization_method="estimated").fit(
	//     smoothing_level=0.8, smoothing_trend=0.2, optimized=False
	// )
	values := []float64{
		17.5534,
		21.86,
		23.8866,
		26.9293,
		26.8885,
		28.8314,
		30.0751,
		30.9535,
		30.1857,
		31.5797,
		32.5776,
		33.4774,
		39.0216,
		41.3864,
		41.5966,
	}
	start := time.Date(1990, 12, 31, 0, 0, 0, 0, time.UTC)
	interval := 365 * 24 * time.Hour
	index := generator.MakeSeriesIndex(start, interval, len(values))

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	forecastHorizon := 5
	alpha := 0.8
	beta := 0.2
	forecast := DoubleExponentialSmoothingEstimated(ts, alpha, beta, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Fatalf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	expectedValues := []float64{
		43.62500632857281,
		45.36318287544541,
		47.10135942231801,
		48.839535969190614,
		50.577712516063215,
	}
	const epsilon = 1e-9
	lastPoint, err := ts.Last()
	if err != nil {
		t.Fatalf("Unexpected error getting last point: %v", err)
	}

	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if math.Abs(dp.Value-expectedValues[i]) > epsilon {
			t.Errorf("At index %d: expected value %f, got %f", i, expectedValues[i], dp.Value)
		}
	}
}

func TestTripleExponentialSmoothingHandCalculated(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Hour
	values := []float64{
		3, 7,
		5, 9,
		7, 11,
	}
	index := generator.MakeSeriesIndex(start, interval, len(values))

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	forecastHorizon := 2
	forecast := TripleExponentialSmoothing(ts, 0.5, 0.5, 0.5, 2, forecastHorizon)

	if forecast.Length() != forecastHorizon {
		t.Fatalf("Expected forecast length %d, got %d", forecastHorizon, forecast.Length())
	}

	expectedValues := []float64{8.79296875, 13.14453125}
	lastPoint, err := ts.Last()
	if err != nil {
		t.Fatalf("Unexpected error getting last point: %v", err)
	}

	for i, dp := range forecast.DataPoints() {
		expectedTime := lastPoint.Timestamp.Add(time.Duration(i+1) * interval)
		if !dp.Timestamp.Equal(expectedTime) {
			t.Errorf("At index %d: expected timestamp %v, got %v", i, expectedTime, dp.Timestamp)
		}
		if math.Abs(dp.Value-expectedValues[i]) > 1e-9 {
			t.Errorf("At index %d: expected value %f, got %f", i, expectedValues[i], dp.Value)
		}
	}
}

func TestTripleExponentialSmoothingPureSeasonality(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	interval := time.Hour
	values := []float64{
		10, 20, 30, 40,
		10, 20, 30, 40,
		10, 20, 30, 40,
	}
	index := generator.MakeSeriesIndex(start, interval, len(values))

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	forecast := TripleExponentialSmoothing(ts, 1.0, 1.0, 1.0, 4, 4)
	expectedValues := []float64{10, 20, 30, 40}

	for i, dp := range forecast.DataPoints() {
		if math.Abs(dp.Value-expectedValues[i]) > 1e-9 {
			t.Errorf("At index %d: expected value %f, got %f", i, expectedValues[i], dp.Value)
		}
	}
}

func TestTripleExponentialSmoothingInvalidInputReturnsEmpty(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := generator.MakeSeriesIndex(start, time.Hour, 6)
	values := []float64{1, 2, 3, 4, 5, 6}

	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		t.Fatalf("Unexpected error creating time series: %v", err)
	}

	cases := []timeseriesgo.TimeSeries{
		TripleExponentialSmoothing(timeseriesgo.Empty(), 0.5, 0.3, 0.2, 3, 2),
		TripleExponentialSmoothing(ts, -0.1, 0.3, 0.2, 3, 2),
		TripleExponentialSmoothing(ts, 0.5, 1.1, 0.2, 3, 2),
		TripleExponentialSmoothing(ts, 0.5, 0.3, 1.2, 3, 2),
		TripleExponentialSmoothing(ts, 0.5, 0.3, 0.2, 0, 2),
		TripleExponentialSmoothing(ts, 0.5, 0.3, 0.2, 4, 2),
		TripleExponentialSmoothing(ts, 0.5, 0.3, 0.2, 3, 0),
	}

	for i, result := range cases {
		if !result.IsEmpty() {
			t.Errorf("Case %d: expected empty forecast, got length %d", i, result.Length())
		}
	}
}
