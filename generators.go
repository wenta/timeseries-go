package timeseriesgo

import (
	"math/rand/v2"
	"time"
)

/**
 * Creates a slice of time.Time representing a series of timestamps.
 *
 * @param start The starting time.Time for the series.
 * @param interval The duration between consecutive timestamps.
 * @param count The number of timestamps to generate.
 *
 * @return A slice of time.Time with the specified number of timestamps.
 */
func MakeSeriesIndex(start time.Time, interval time.Duration, count int) []time.Time {
	ts := []time.Time{}
	for i := 0; i < count; i++ {
		ts = append(ts, start.Add(time.Duration(i)*interval))
	}
	return ts
}

/**
 * Generates a TimeSeries with constant value at specified timestamps.
 *
 * @param index A slice of time.Time representing the timestamps for the DataPoints.
 * @param value The constant value for each DataPoint.
 *
 * @return A TimeSeries with DataPoints at the specified timestamps, all having the same value.
 */
func Constant(index []time.Time, value float64) TimeSeries {
	ts := Empty()
	for _, dt := range index {
		ts.AddPoint(DataPoint{
			timestamp: dt,
			value:     value,
		})
	}
	return ts
}

/**
 * Generates a TimeSeries representing a random walk starting from a given value.
 *
 * @param index A slice of time.Time representing the timestamps for the DataPoints.
 * @param startValue The starting value for the random walk.
 *
 * @return A TimeSeries with DataPoints at the specified timestamps, where each value is derived from the previous one by adding or subtracting 1.0 randomly.
 */

func RandomWalk(index []time.Time, startValue float64) TimeSeries {
	ts := Empty()
	nextValue := startValue
	for _, dt := range index {
		ts.AddPoint(DataPoint{
			timestamp: dt,
			value:     nextValue,
		})
		if rand.IntN(2) == 0 {
			nextValue -= 1.0
		} else {
			nextValue += 1.0
		}
	}
	return ts
}

func Repeat(pattern TimeSeries, start time.Time, end time.Time) TimeSeries {
	if pattern.IsEmpty() {
		return Empty()
	} else {
		ts := Empty()
		resolution, err := pattern.Resolution()
		if err != nil {
			return pattern
		}
		i := 0
		vs := pattern.Values()
		for now := start; now.Before(end); now = now.Add(resolution) {
			if i == pattern.Length() {
				i = 0
			}
			ts.AddPoint(DataPoint{now, vs[i]})
			i++
		}
		return ts
	}
}
