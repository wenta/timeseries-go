package anomaly

import (
	"errors"
	"math"

	timeseriesgo "github.com/wenta/timeseries-go"
)

/**
 * Calculates the Z-Score normalization of the TimeSeries.
 *
 * @param ts The TimeSeries to be normalized.
 *
 * @return A new TimeSeries with Z-Score normalized values, or an error if the calculation fails.
 */
func ZScore(ts timeseriesgo.TimeSeries) (timeseriesgo.TimeSeries, error) {
	if ts.IsEmpty() {
		return timeseriesgo.Empty(), nil
	}

	points := ts.DataPoints()
	return zScoreSeries(points)
}

/**
 * Finds anomalies in the TimeSeries using Z-Score thresholding.
 *
 * @param ts The TimeSeries to be analyzed.
 *
 * @return A new TimeSeries with 1 for anomalous points and 0 for normal points, or an error if the calculation fails.
 */
func FindAnomaliesWithZScore(ts timeseriesgo.TimeSeries) (timeseriesgo.TimeSeries, error) {
	if ts.IsEmpty() {
		return timeseriesgo.Empty(), nil
	}

	points := ts.DataPoints()
	return zScoreFlags(points, 2)
}

/**
 * Calculates the Robust Z-Score normalization of the TimeSeries.
 *
 * @param ts The TimeSeries to be normalized.
 *
 * @return A new TimeSeries with robust Z-Score normalized values, or an error if the calculation fails.
 */
func RobustZScore(ts timeseriesgo.TimeSeries) (timeseriesgo.TimeSeries, error) {
	if ts.IsEmpty() {
		return timeseriesgo.Empty(), errors.New("timeseries is empty")
	}

	points := ts.DataPoints()
	return robustZScoreSeries(points)
}

/**
 * Finds anomalies in the TimeSeries using Robust Z-Score thresholding.
 *
 * @param ts The TimeSeries to be analyzed.
 *
 * @return A new TimeSeries with 1 for anomalous points and 0 for normal points, or an error if the calculation fails.
 */
func FindAnomaliesWithRobustZScore(ts timeseriesgo.TimeSeries) (timeseriesgo.TimeSeries, error) {
	if ts.IsEmpty() {
		return timeseriesgo.Empty(), errors.New("timeseries is empty")
	}

	points := ts.DataPoints()
	return robustZScoreFlags(points, 3)
}

/**
 * Finds spike anomalies in the TimeSeries based on a positive jump threshold.
 *
 * @param ts The TimeSeries to be analyzed.
 * @param threshold The minimum positive jump required to mark a spike anomaly.
 *
 * @return A new TimeSeries with 1 for spike anomalies and 0 for normal points, or an error if the input is invalid.
 */
func FindSpikeAnomalies(ts timeseriesgo.TimeSeries, threshold float64) (timeseriesgo.TimeSeries, error) {
	if ts.IsEmpty() {
		return timeseriesgo.Empty(), errors.New("timeseries is empty")
	}
	if threshold <= 0 {
		return timeseriesgo.Empty(), errors.New("spike threshold must be positive")
	}

	flags := make([]float64, ts.Length())
	points := ts.DataPoints()

	for i := 1; i < ts.Length(); i++ {
		diff := points[i].Value - points[i-1].Value
		if diff >= threshold {
			flags[i] = 1
		}
	}
	return seriesFromValues(points, flags), nil
}

/**
 * Finds drop anomalies in the TimeSeries based on a negative jump threshold.
 *
 * @param ts The TimeSeries to be analyzed.
 * @param threshold The minimum drop magnitude required to mark a drop anomaly.
 *
 * @return A new TimeSeries with 1 for drop anomalies and 0 for normal points, or an error if the input is invalid.
 */
func FindDropAnomalies(ts timeseriesgo.TimeSeries, threshold float64) (timeseriesgo.TimeSeries, error) {
	if ts.IsEmpty() {
		return timeseriesgo.Empty(), errors.New("timeseries is empty")
	}
	if threshold <= 0 {
		return timeseriesgo.Empty(), errors.New("drop threshold must be positive")
	}

	flags := make([]float64, ts.Length())
	points := ts.DataPoints()
	for i := 1; i < ts.Length(); i++ {
		diff := points[i].Value - points[i-1].Value
		if diff <= -threshold {
			flags[i] = 1
		}
	}
	return seriesFromValues(points, flags), nil
}

/**
 * Finds flatline anomalies in the TimeSeries based on tolerance and minimum run length.
 *
 * @param ts The TimeSeries to be analyzed.
 * @param tolerance The maximum allowed difference between adjacent values in a flatline.
 * @param minLength The minimum number of consecutive points required to mark a flatline.
 *
 * @return A new TimeSeries with 1 for flatline anomalies and 0 for normal points, or an error if the input is invalid.
 */
func FindFlatlineAnomalies(ts timeseriesgo.TimeSeries, tolerance float64, minLength int) (timeseriesgo.TimeSeries, error) {
	if ts.IsEmpty() {
		return timeseriesgo.Empty(), errors.New("timeseries is empty")
	}
	if tolerance < 0 {
		return timeseriesgo.Empty(), errors.New("flatline tolerance must be non-negative")
	}
	if minLength <= 0 {
		return timeseriesgo.Empty(), errors.New("flatline minimum length must be positive")
	}

	flags := make([]float64, ts.Length())
	runStart := 0
	runLength := 1
	points := ts.DataPoints()

	for i := 1; i < ts.Length(); i++ {
		delta := math.Abs(points[i].Value - points[i-1].Value)
		if delta <= tolerance {
			runLength++
		} else {
			if runLength >= minLength {
				for j := runStart; j < runStart+runLength; j++ {
					flags[j] = 1
				}
			}
			runStart = i
			runLength = 1
		}
	}

	if runLength >= minLength {
		for j := runStart; j < runStart+runLength; j++ {
			flags[j] = 1
		}
	}
	return seriesFromValues(points, flags), nil
}

func zScoreSeries(points []timeseriesgo.DataPoint) (timeseriesgo.TimeSeries, error) {
	mean, sampleVariance := meanAndSampleVariance(points)
	stddev := math.Sqrt(sampleVariance)
	values := make([]float64, len(points))
	for i, dp := range points {
		values[i] = (dp.Value - mean) / stddev
	}
	return seriesFromValues(points, values), nil
}

func zScoreFlags(points []timeseriesgo.DataPoint, threshold float64) (timeseriesgo.TimeSeries, error) {
	mean, sampleVariance := meanAndSampleVariance(points)
	stddev := math.Sqrt(sampleVariance)
	flags := make([]float64, len(points))
	for i, dp := range points {
		if math.Abs((dp.Value-mean)/stddev) > threshold {
			flags[i] = 1
		}
	}
	return seriesFromValues(points, flags), nil
}

func robustZScoreSeries(points []timeseriesgo.DataPoint) (timeseriesgo.TimeSeries, error) {
	median, err := pointMedian(points)
	if err != nil {
		return timeseriesgo.Empty(), err
	}

	deviations := make([]float64, len(points))
	for i, dp := range points {
		deviations[i] = math.Abs(dp.Value - median)
	}

	mad, err := medianFromValues(deviations)
	if err != nil {
		return timeseriesgo.Empty(), err
	}

	scaledMAD := mad * 1.4826
	values := make([]float64, len(points))
	for i, dp := range points {
		values[i] = (dp.Value - median) / scaledMAD
	}
	return seriesFromValues(points, values), nil
}

func robustZScoreFlags(points []timeseriesgo.DataPoint, threshold float64) (timeseriesgo.TimeSeries, error) {
	median, err := pointMedian(points)
	if err != nil {
		return timeseriesgo.Empty(), err
	}

	deviations := make([]float64, len(points))
	for i, dp := range points {
		deviations[i] = math.Abs(dp.Value - median)
	}

	mad, err := medianFromValues(deviations)
	if err != nil {
		return timeseriesgo.Empty(), err
	}

	scaledMAD := mad * 1.4826
	flags := make([]float64, len(points))
	for i, dp := range points {
		if math.Abs((dp.Value-median)/scaledMAD) > threshold {
			flags[i] = 1
		}
	}
	return seriesFromValues(points, flags), nil
}

func meanAndSampleVariance(points []timeseriesgo.DataPoint) (float64, float64) {
	sum := 0.0
	for _, point := range points {
		sum += point.Value
	}

	mean := sum / float64(len(points))
	sumSquares := 0.0
	for _, point := range points {
		diff := point.Value - mean
		sumSquares += diff * diff
	}

	if len(points) == 1 {
		return mean, sumSquares
	}
	return mean, sumSquares / float64(len(points)-1)
}

func pointMedian(points []timeseriesgo.DataPoint) (float64, error) {
	values := make([]float64, len(points))
	for i, dp := range points {
		values[i] = dp.Value
	}
	return medianFromValues(values)
}

func medianFromValues(values []float64) (float64, error) {
	if len(values) == 0 {
		return 0, errors.New("timeseries is empty")
	}

	position := float64(50*(len(values)+1)) / 100
	if position < 1 {
		return values[0], nil
	}
	if position >= float64(len(values)) {
		return values[len(values)-1], nil
	}

	lowerIndex := int(math.Floor(position)) - 1
	upperIndex := lowerIndex + 1
	fraction := position - math.Floor(position)
	return values[lowerIndex] + fraction*(values[upperIndex]-values[lowerIndex]), nil
}

func seriesFromValues(points []timeseriesgo.DataPoint, values []float64) timeseriesgo.TimeSeries {
	result := make([]timeseriesgo.DataPoint, len(points))
	for i, dp := range points {
		result[i] = timeseriesgo.DataPoint{
			Timestamp: dp.Timestamp,
			Value:     values[i],
		}
	}
	return timeseriesgo.FromDataPoints(result)
}
