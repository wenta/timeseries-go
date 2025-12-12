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

// FindSpikeAnomalies flags positive jumps greater than or equal to the given threshold.
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

// FindDropAnomalies flags negative jumps (drops) greater than or equal to the given threshold.
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

// FindFlatlineAnomalies flags runs of near-constant values.
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
