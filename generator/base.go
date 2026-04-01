package generator

import (
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

/**
 * Generates a TimeSeries with constant value at specified timestamps.
 *
 * @param index A slice of time.Time representing the timestamps for the DataPoints.
 * @param value The constant value for each DataPoint.
 *
 * @return A TimeSeries with DataPoints at the specified timestamps, all having the same value.
 */
func Constant(index []time.Time, value float64) timeseriesgo.TimeSeries {
	ts := timeseriesgo.Empty()
	for _, dt := range index {
		ts.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dt,
			Value:     value,
		})
	}
	return ts
}

/**
 * Repeats a pattern TimeSeries between start and end using the pattern resolution.
 *
 * @param pattern The pattern TimeSeries to repeat.
 * @param start The inclusive start timestamp of the generated range.
 * @param end The exclusive end timestamp of the generated range.
 *
 * @return A TimeSeries repeating the pattern over [start, end), or an empty series when the pattern is empty.
 */
func Repeat(pattern timeseriesgo.TimeSeries, start time.Time, end time.Time) timeseriesgo.TimeSeries {
	if pattern.IsEmpty() {
		return timeseriesgo.Empty()
	}

	ts := timeseriesgo.Empty()
	resolution, err := pattern.Resolution()
	if err != nil {
		return pattern
	}

	values := pattern.Values()
	position := 0
	for now := start; now.Before(end); now = now.Add(resolution) {
		if position == pattern.Length() {
			position = 0
		}
		ts.AddPoint(timeseriesgo.DataPoint{
			Timestamp: now,
			Value:     values[position],
		})
		position++
	}

	return ts
}
