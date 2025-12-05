package timeseriesgo

import "errors"

type MeanAndVariance struct {
	mean               float64
	sampleVariance     float64
	populationVariance float64
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
			mean:               mean,
			sampleVariance:     sampleVariance,
			populationVariance: populationVariance,
		}, nil
	}
}
