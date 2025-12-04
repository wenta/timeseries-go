package timeseriesgo

import (
	"errors"
	"math"
)

/**
 * Calculates the Mean Squared Error (MSE) between two TimeSeries.
 *
 * @param ts1 The first TimeSeries.
 * @param ts2 The second TimeSeries.
 *
 * @return The MSE value, or an error if either TimeSeries is empty.
 */
func MSE(ts1, ts2 TimeSeries) (float64, error) {
	if ts1.IsEmpty() || ts2.IsEmpty() {
		return 0.0, errors.New("one or both TimeSeries are empty")
	} else {
		joined := ts1.Join(ts2)
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
func RMSE(ts1, ts2 TimeSeries) (float64, error) {
	mse, err := MSE(ts1, ts2)
	if err != nil {
		return 0.0, err
	} else {
		return math.Sqrt(mse), nil
	}
}
