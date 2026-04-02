package stats

import (
	"errors"
	"math"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

type MeanAndVariance struct {
	Mean               float64
	SampleVariance     float64
	PopulationVariance float64
}

// SeriesComparison stores a small comparison summary for original and synthetic series.
type SeriesComparison struct {
	LengthOriginal    int
	LengthSynthetic   int
	MeanOriginal      float64
	MeanSynthetic     float64
	VarianceOriginal  float64
	VarianceSynthetic float64
	Lag1Original      float64
	Lag1Synthetic     float64
}

/**
 * Calculates the mean and variance of the values in the TimeSeries.
 *
 * @param ts The TimeSeries whose values are used to compute summary statistics.
 *
 * @return A MeanAndVariance struct containing mean, sample variance, and population variance, or an error if the series is empty.
 */
func GetMeanAndVariance(ts timeseriesgo.TimeSeries) (MeanAndVariance, error) {
	if ts.IsEmpty() {
		return MeanAndVariance{}, errors.New("TimeSeries is empty")
	}

	mean := ts.Sum() / float64(ts.Length())
	sampleVariance := 0.0
	for _, v := range ts.Values() {
		diff := v - mean
		sampleVariance += diff * diff
	}
	populationVariance := sampleVariance / float64(ts.Length())
	// Use sample variance (divide by n-1) to avoid underestimating stddev on small samples.
	if ts.Length() > 1 {
		sampleVariance /= float64(ts.Length() - 1)
	}
	return MeanAndVariance{
		Mean:               mean,
		SampleVariance:     sampleVariance,
		PopulationVariance: populationVariance,
	}, nil
}

/**
 * Calculates the moving average of the TimeSeries over a trailing time window.
 *
 * @param ts The TimeSeries to be averaged.
 * @param window The trailing time window used to compute each average.
 *
 * @return A new TimeSeries containing moving-average values at the original timestamps.
 */
func MovingAverage(ts timeseriesgo.TimeSeries, window time.Duration) timeseriesgo.TimeSeries {
	if ts.IsEmpty() {
		return timeseriesgo.Empty()
	}

	if window <= 0 {
		cloned := ts.DataPoints()
		return timeseriesgo.FromDataPoints(cloned)
	}

	result := timeseriesgo.Empty()
	left := 0
	runningSum := 0.0
	points := ts.DataPoints()

	for right, dp := range points {
		runningSum += dp.Value

		// Maintain window (t-window, t] to match RollingWindow semantics.
		for left <= right && dp.Timestamp.Sub(points[left].Timestamp) >= window {
			runningSum -= points[left].Value
			left++
		}

		count := right - left + 1
		result.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dp.Timestamp,
			Value:     runningSum / float64(count),
		})
	}

	return result
}

/**
 * Calculates the Pearson correlation coefficient between two TimeSeries.
 *
 * @param ts1 The first TimeSeries to compare.
 * @param ts2 The second TimeSeries to compare.
 *
 * @return The Pearson correlation coefficient computed on overlapping timestamps, or an error if the calculation cannot be performed.
 */
func Correlation(ts1, ts2 timeseriesgo.TimeSeries) (float64, error) {
	if ts1.IsEmpty() || ts2.IsEmpty() {
		return 0, errors.New("one or both TimeSeries are empty")
	}

	points1 := ts1.DataPoints()
	points2 := ts2.DataPoints()
	left := 0
	right := 0
	count := 0
	meanX := 0.0
	meanY := 0.0
	sumSquaresX := 0.0
	sumSquaresY := 0.0
	sumProducts := 0.0

	for left < len(points1) && right < len(points2) {
		switch points1[left].Timestamp.Compare(points2[right].Timestamp) {
		case -1:
			left++
		case 1:
			right++
		default:
			count++
			x := points1[left].Value
			y := points2[right].Value
			deltaX := x - meanX
			meanX += deltaX / float64(count)
			deltaY := y - meanY
			meanY += deltaY / float64(count)

			sumSquaresX += deltaX * (x - meanX)
			sumSquaresY += deltaY * (y - meanY)
			sumProducts += deltaX * (y - meanY)

			left++
			right++
		}
	}

	if count < 2 {
		return 0, errors.New("not enough aligned points")
	}
	if sumSquaresX == 0 || sumSquaresY == 0 {
		return 0, errors.New("zero variance in one of the series")
	}

	return sumProducts / math.Sqrt(sumSquaresX*sumSquaresY), nil
}

/**
 * Calculates the autocorrelation function (ACF) of a TimeSeries for lags 0..nlags.
 *
 * The returned slice always starts with lag 0, which equals 1 for a series with non-zero variance.
 * This implementation uses the full-series mean and variance denominator, matching the
 * standard biased autocorrelation estimator.
 *
 * @param ts The TimeSeries whose autocorrelation is computed.
 * @param nlags The maximum lag to include in the result.
 *
 * @return A slice of autocorrelation values for lags 0 through nlags, or an error if the calculation cannot be performed.
 */
func ACF(ts timeseriesgo.TimeSeries, nlags int) ([]float64, error) {
	if ts.IsEmpty() {
		return nil, errors.New("TimeSeries is empty")
	}
	if nlags < 0 {
		return nil, errors.New("nlags must be non-negative")
	}
	if ts.Length() < 2 {
		return nil, errors.New("TimeSeries must contain at least two points")
	}
	if nlags >= ts.Length() {
		return nil, errors.New("nlags must be smaller than the series length")
	}

	values := ts.Values()
	mean := ts.Sum() / float64(ts.Length())

	denominator := 0.0
	centered := make([]float64, len(values))
	for i, value := range values {
		centered[i] = value - mean
		denominator += centered[i] * centered[i]
	}

	if denominator == 0 {
		return nil, errors.New("zero variance in TimeSeries")
	}

	acf := make([]float64, nlags+1)
	for lag := 0; lag <= nlags; lag++ {
		numerator := 0.0
		for i := lag; i < len(centered); i++ {
			numerator += centered[i] * centered[i-lag]
		}
		acf[lag] = numerator / denominator
	}

	return acf, nil
}

/**
 * Calculates autocorrelation for a single lag of a TimeSeries.
 *
 * This helper uses the same estimator as ACF and returns the value at the
 * requested lag directly.
 *
 * @param ts The TimeSeries whose autocorrelation is computed.
 * @param lag The lag to evaluate. Lag 0 equals 1 for a series with non-zero variance.
 *
 * @return The autocorrelation at the requested lag, or an error if the calculation cannot be performed.
 */
func Autocorrelation(ts timeseriesgo.TimeSeries, lag int) (float64, error) {
	acf, err := ACF(ts, lag)
	if err != nil {
		return 0, err
	}
	return acf[lag], nil
}

/**
 * Compares two series using stable summary statistics on raw values.
 *
 * The comparison is computed independently for each series without aligning
 * timestamps. Variance fields use the population variance returned by
 * GetMeanAndVariance to keep the summary stable across different lengths.
 * Lag-1 autocorrelation is included when it can be computed; otherwise the
 * corresponding field is set to NaN.
 *
 * @param original The reference TimeSeries.
 * @param synthetic The synthetic or derived TimeSeries to compare.
 *
 * @return A SeriesComparison with mean and variance summaries for both series, or an error if either series is empty.
 */
func CompareSeriesStats(original, synthetic timeseriesgo.TimeSeries) (SeriesComparison, error) {
	if original.IsEmpty() || synthetic.IsEmpty() {
		return SeriesComparison{}, errors.New("one or both TimeSeries are empty")
	}

	originalStats, err := GetMeanAndVariance(original)
	if err != nil {
		return SeriesComparison{}, err
	}

	syntheticStats, err := GetMeanAndVariance(synthetic)
	if err != nil {
		return SeriesComparison{}, err
	}

	return SeriesComparison{
		LengthOriginal:    original.Length(),
		LengthSynthetic:   synthetic.Length(),
		MeanOriginal:      originalStats.Mean,
		MeanSynthetic:     syntheticStats.Mean,
		VarianceOriginal:  originalStats.PopulationVariance,
		VarianceSynthetic: syntheticStats.PopulationVariance,
		Lag1Original:      lagOrNaN(original),
		Lag1Synthetic:     lagOrNaN(synthetic),
	}, nil
}

func lagOrNaN(ts timeseriesgo.TimeSeries) float64 {
	if ts.Length() < 2 {
		return math.NaN()
	}

	value, err := Autocorrelation(ts, 1)
	if err != nil {
		return math.NaN()
	}
	return value
}

/**
 * Normalizes the TimeSeries values to the [0,1] range while preserving timestamps.
 *
 * @param ts The TimeSeries to be normalized.
 *
 * @return A new normalized TimeSeries, or an error if the series is empty.
 */
func MinMaxNormalize(ts timeseriesgo.TimeSeries) (timeseriesgo.TimeSeries, error) {
	if ts.IsEmpty() {
		return timeseriesgo.Empty(), errors.New("TimeSeries is empty")
	}

	points := ts.DataPoints()

	minVal := points[0].Value
	maxVal := points[0].Value

	// Find min and max
	for _, dp := range points {
		if dp.Value < minVal {
			minVal = dp.Value
		}
		if dp.Value > maxVal {
			maxVal = dp.Value
		}
	}

	result := timeseriesgo.Empty()

	// Avoid division by zero when all values are equal
	if maxVal == minVal {
		for _, dp := range points {
			result.AddPoint(timeseriesgo.DataPoint{
				Timestamp: dp.Timestamp,
				Value:     0.0,
			})
		}
		return result, nil
	}

	// Apply normalization
	denominator := maxVal - minVal
	for _, dp := range points {
		normalized := (dp.Value - minVal) / denominator
		result.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dp.Timestamp,
			Value:     normalized,
		})
	}

	return result, nil
}
