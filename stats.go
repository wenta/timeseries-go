package timeseriesgo

import "errors"

type MeanAndVariance struct {
	mean     float64
	variance float64
}

/**
 * Calculates the mean and variance of the values in the TimeSeries.
 *
 * @return A MeanAndVariance struct containing the mean and variance, or an error if the TimeSeries is empty.
 */
func (ts *TimeSeries) GetMeanAndVariance() (MeanAndVariance, error) {
	if ts.IsEmpty() {
		return MeanAndVariance{}, errors.New("TimeSeries is empty")
	} else {
		mean := ts.Sum() / float64(ts.Length())
		variance := 0.0
		for _, v := range ts.Values() {
			diff := v - mean
			variance += diff * diff
		}
		variance /= float64(ts.Length())
		return MeanAndVariance{mean: mean, variance: variance}, nil
	}
}
