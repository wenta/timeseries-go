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
