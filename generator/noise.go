package generator

import (
	"math/rand/v2"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

/**
 * Generates a TimeSeries representing a random walk starting from a given value.
 *
 * @param index A slice of time.Time representing the timestamps for the DataPoints.
 * @param startValue The starting value for the random walk.
 *
 * @return A TimeSeries with DataPoints at the specified timestamps, where each value is derived from the previous one by adding or subtracting 1.0 randomly.
 */
func RandomWalk(index []time.Time, startValue float64) timeseriesgo.TimeSeries {
	ts := timeseriesgo.Empty()
	nextValue := startValue
	for _, dt := range index {
		ts.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dt,
			Value:     nextValue,
		})
		if rand.IntN(2) == 0 {
			nextValue -= 1.0
		} else {
			nextValue += 1.0
		}
	}
	return ts
}

/**
 * Generates a TimeSeries containing Gaussian random noise.
 *
 * @param index A slice of time.Time representing the timestamps for the DataPoints.
 * @param mean The expected mean of the generated values.
 * @param stddev The standard deviation of the generated values.
 *
 * @return A TimeSeries with DataPoints at the specified timestamps, where each
 *         value is sampled independently from N(mean, stddev^2).
 */
func RandomNoise(index []time.Time, mean float64, stddev float64) timeseriesgo.TimeSeries {
	ts := timeseriesgo.Empty()
	for _, dt := range index {
		ts.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dt,
			Value:     mean + stddev*rand.NormFloat64(),
		})
	}
	return ts
}

/**
 * Generates a TimeSeries containing uniformly distributed random noise.
 *
 * @param index A slice of time.Time representing the timestamps for the DataPoints.
 * @param min The lower bound of the generated values.
 * @param max The upper bound of the generated values.
 *
 * @return A TimeSeries with DataPoints at the specified timestamps, where each
 *         value is sampled independently from a uniform distribution over [min, max].
 */
func UniformNoise(index []time.Time, min float64, max float64) timeseriesgo.TimeSeries {
	if max < min {
		min, max = max, min
	}

	ts := timeseriesgo.Empty()
	for _, dt := range index {
		value := min
		if max != min {
			value += rand.Float64() * (max - min)
		}
		ts.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dt,
			Value:     value,
		})
	}
	return ts
}
