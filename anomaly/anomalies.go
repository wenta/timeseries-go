package anomaly

import (
	"errors"
	"math"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/stats"
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

	mv, err := stats.GetMeanAndVariance(ts)
	if err != nil {
		return timeseriesgo.Empty(), err
	}
	mean := mv.Mean
	stddev := math.Sqrt(mv.SampleVariance)
	zscored := timeseriesgo.Empty()
	for _, dp := range ts.DataPoints() {
		zscored.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dp.Timestamp,
			Value:     (dp.Value - mean) / stddev,
		})
	}
	return zscored, nil
}

/**
 * Finds anomalies in the TimeSeries using Z-Score thresholding.
 *
 * @param ts The TimeSeries to be analyzed.
 *
 * @return A new TimeSeries with 1 for anomalous points and 0 for normal points, or an error if the calculation fails.
 */
func FindAnomaliesWithZScore(ts timeseriesgo.TimeSeries) (timeseriesgo.TimeSeries, error) {
	rs, err := ZScore(ts)
	if err != nil {
		return timeseriesgo.Empty(), err
	} else {
		return rs.MapValues(func(x float64) float64 {
			if math.Abs(x) > 2 {
				return 1
			} else {
				return 0
			}
		}), nil
	}
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
	} else {
		median, err := ts.Median()
		if err != nil {
			return timeseriesgo.Empty(), err
		}
		deviations := ts.MapValues(func(x float64) float64 {
			return math.Abs(x - median)
		})
		mad, err2 := deviations.Median()
		if err2 != nil {
			return timeseriesgo.Empty(), err
		}
		scaledMAD := mad * 1.4826
		return ts.MapValues(func(x float64) float64 {
			return (x - median) / scaledMAD
		}), nil
	}
}

/**
 * Finds anomalies in the TimeSeries using Robust Z-Score thresholding.
 *
 * @param ts The TimeSeries to be analyzed.
 *
 * @return A new TimeSeries with 1 for anomalous points and 0 for normal points, or an error if the calculation fails.
 */
func FindAnomaliesWithRobustZScore(ts timeseriesgo.TimeSeries) (timeseriesgo.TimeSeries, error) {
	rs, err := RobustZScore(ts)
	if err != nil {
		return timeseriesgo.Empty(), err
	} else {
		return rs.MapValues(func(x float64) float64 {
			if math.Abs(x) > 3 {
				return 1
			} else {
				return 0
			}
		}), nil
	}
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

	res := timeseriesgo.Empty()
	for i, dp := range points {
		res.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dp.Timestamp,
			Value:     flags[i],
		})
	}
	return res, nil
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

	res := timeseriesgo.Empty()
	for i, dp := range points {
		res.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dp.Timestamp,
			Value:     flags[i],
		})
	}
	return res, nil
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

	res := timeseriesgo.Empty()
	for i, dp := range points {
		res.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dp.Timestamp,
			Value:     flags[i],
		})
	}
	return res, nil
}
