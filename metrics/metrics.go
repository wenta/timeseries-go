package metrics

import (
	"errors"
	"math"
	"slices"

	timeseriesgo "github.com/wenta/timeseries-go"
)

/**
 * Calculates the Mean Squared Error (MSE) between two TimeSeries.
 *
 * @param ts1 The first TimeSeries.
 * @param ts2 The second TimeSeries.
 *
 * @return The MSE value, or an error if either TimeSeries is empty.
 */
func MSE(ts1, ts2 timeseriesgo.TimeSeries) (float64, error) {
	if ts1.IsEmpty() || ts2.IsEmpty() {
		return 0.0, errors.New("one or both TimeSeries are empty")
	} else {
		joined := ts1.Join(ts2)
		if joined.Length() == 0 {
			return 0.0, errors.New("no overlapping timestamps between series")
		}
		ts := joined.MapValuesWithReduce(func(l, r float64) float64 {
			diff := l - r
			return diff * diff
		})
		return ts.Sum() / float64(ts.Length()), nil
	}
}

/**
 * Calculates the Root Mean Squared Error (RMSE) between two TimeSeries.
 *
 * @param ts1 The first TimeSeries.
 * @param ts2 The second TimeSeries.
 *
 * @return The RMSE value, or an error if either TimeSeries is empty.
 */
func RMSE(ts1, ts2 timeseriesgo.TimeSeries) (float64, error) {
	mse, err := MSE(ts1, ts2)
	if err != nil {
		return 0.0, err
	} else {
		return math.Sqrt(mse), nil
	}
}

/**
 * Calculates the Mean Absolute Error (MAE) between two TimeSeries.
 *
 * @param ts1 The first TimeSeries.
 * @param ts2 The second TimeSeries.
 *
 * @return The MAE value, or an error if either TimeSeries is empty.
 */
func MAE(ts1, ts2 timeseriesgo.TimeSeries) (float64, error) {
	if ts1.IsEmpty() || ts2.IsEmpty() {
		return 0.0, errors.New("one or both TimeSeries are empty")
	} else {
		joined := ts1.Join(ts2)
		if joined.Length() == 0 {
			return 0.0, errors.New("no overlapping timestamps between series")
		}
		ts := joined.MapValuesWithReduce(func(l, r float64) float64 {
			return math.Abs(l - r)
		})
		vs := ts.Values()
		slices.Sort(vs)

		if ts.Length()%2 == 0 {
			mid1 := vs[(ts.Length()/2)-1]
			mid2 := vs[ts.Length()/2]
			return (mid1 + mid2) / 2.0, nil
		} else {
			mid := vs[ts.Length()/2]
			return mid, nil

		}
	}
}

func MAD(ts timeseriesgo.TimeSeries) (float64, error) {
	if ts.IsEmpty() {
		return 0.0, errors.New("timeseries is empty")
	} else {
		median, err := ts.Median()
		if err != nil {
			return 0.0, err
		}
		deviations := ts.MapValues(func(x float64) float64 {
			return math.Abs(x - median)
		})
		return deviations.Median()
	}
}
