package timeseriesgo

import (
	"errors"
	"math"
)

/**
 * Calculates the Z-Score normalization of the TimeSeries.
 *
 * @param ts The TimeSeries to be normalized.
 *
 * @return A new TimeSeries with Z-Score normalized values, or an error if the calculation fails.
 */
func ZScore(ts TimeSeries) (TimeSeries, error) {
	if ts.IsEmpty() {
		return Empty(), nil
	} else {
		mv, err := ts.GetMeanAndVariance()
		if err != nil {
			return Empty(), err
		}
		mean := mv.mean
		stddev := math.Sqrt(mv.sampleVariance)
		zscored := Empty()
		for _, dp := range ts.datapoints {
			zscored.AddPoint(DataPoint{
				timestamp: dp.timestamp,
				value:     (dp.value - mean) / stddev,
			})
		}
		return zscored, nil
	}
}

func FindAnomaliesWithZScore(ts TimeSeries) (TimeSeries, error) {
	rs, err := ZScore(ts)
	if err != nil {
		return Empty(), err
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

func RobustZScore(ts TimeSeries) (TimeSeries, error) {
	if ts.IsEmpty() {
		return Empty(), errors.New("timeseries is empty")
	} else {
		median, err := ts.Median()
		if err != nil {
			return Empty(), err
		}
		deviations := ts.MapValues(func(x float64) float64 {
			return math.Abs(x - median)
		})
		mad, err2 := deviations.Median()
		if err2 != nil {
			return Empty(), err
		}
		scaledMAD := mad * 1.4826
		return ts.MapValues(func(x float64) float64 {
			return (x - median) / scaledMAD
		}), nil
	}
}

func FindAnomaliesWithRobustZScore(ts TimeSeries) (TimeSeries, error) {
	rs, err := RobustZScore(ts)
	if err != nil {
		return Empty(), err
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
func FindSpikeAnomalies(ts TimeSeries, threshold float64) (TimeSeries, error) {
	if ts.IsEmpty() {
		return Empty(), errors.New("timeseries is empty")
	}
	if threshold <= 0 {
		return Empty(), errors.New("spike threshold must be positive")
	}

	flags := make([]float64, ts.Length())

	for i := 1; i < ts.Length(); i++ {
		diff := ts.datapoints[i].value - ts.datapoints[i-1].value
		if diff >= threshold {
			flags[i] = 1
		}
	}

	res := Empty()
	for i, dp := range ts.datapoints {
		res.AddPoint(DataPoint{
			timestamp: dp.timestamp,
			value:     flags[i],
		})
	}
	return res, nil
}

// FindDropAnomalies flags negative jumps (drops) greater than or equal to the given threshold.
func FindDropAnomalies(ts TimeSeries, threshold float64) (TimeSeries, error) {
	if ts.IsEmpty() {
		return Empty(), errors.New("timeseries is empty")
	}
	if threshold <= 0 {
		return Empty(), errors.New("drop threshold must be positive")
	}

	flags := make([]float64, ts.Length())
	for i := 1; i < ts.Length(); i++ {
		diff := ts.datapoints[i].value - ts.datapoints[i-1].value
		if diff <= -threshold {
			flags[i] = 1
		}
	}

	res := Empty()
	for i, dp := range ts.datapoints {
		res.AddPoint(DataPoint{
			timestamp: dp.timestamp,
			value:     flags[i],
		})
	}
	return res, nil
}

// FindFlatlineAnomalies flags runs of near-constant values.
func FindFlatlineAnomalies(ts TimeSeries, tolerance float64, minLength int) (TimeSeries, error) {
	if ts.IsEmpty() {
		return Empty(), errors.New("timeseries is empty")
	}
	if tolerance < 0 {
		return Empty(), errors.New("flatline tolerance must be non-negative")
	}
	if minLength <= 0 {
		return Empty(), errors.New("flatline minimum length must be positive")
	}

	flags := make([]float64, ts.Length())
	runStart := 0
	runLength := 1

	for i := 1; i < ts.Length(); i++ {
		delta := math.Abs(ts.datapoints[i].value - ts.datapoints[i-1].value)
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

	res := Empty()
	for i, dp := range ts.datapoints {
		res.AddPoint(DataPoint{
			timestamp: dp.timestamp,
			value:     flags[i],
		})
	}
	return res, nil
}
