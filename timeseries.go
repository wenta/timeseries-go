package timeseriesgo

import (
	"errors"
	"fmt"
	"time"
)

type DataPoint struct {
	timestamp time.Time
	value     float64
}

type TimeSeries struct {
	datapoints []DataPoint
}

func Empty() TimeSeries {
	return TimeSeries{datapoints: []DataPoint{}}
}

func (ts *TimeSeries) IsEmpty() bool {
	return len(ts.datapoints) == 0
}

func (ts *TimeSeries) Length() int {
	return len(ts.datapoints)
}

func (ts *TimeSeries) Values() []float64 {
	var res []float64
	for _, dp := range ts.datapoints {
		res = append(res, dp.value)
	}
	return res
}

/**
 * Adds a DataPoint to the TimeSeries.
 *
 * @param dp The DataPoint to add.
 */
func (ts *TimeSeries) AddPoint(dp DataPoint) {
	ts.datapoints = append(ts.datapoints, dp)
}

/**
 * Prints the TimeSeries in a human-readable format.
 */
func (ts *TimeSeries) Print() {
	fmt.Println("Timestamp, Value")
	for _, dp := range ts.datapoints {
		fmt.Printf("%s, %.2f\n", dp.timestamp.Format(time.RFC3339), dp.value)
	}
}

/**
 * Slices the TimeSeries between the specified start and end times.
 *
 * @param start The starting time.Time for the slice (inclusive).
 * @param end The ending time.Time for the slice (exclusive).
 *
 * @return A new TimeSeries containing DataPoints within the specified time range.
 */
func (ts TimeSeries) Slice(start time.Time, end time.Time) TimeSeries {
	sliced := Empty()
	for _, dp := range ts.datapoints {
		if (dp.timestamp.Equal(start) || dp.timestamp.After(start)) && dp.timestamp.Before(end) {
			sliced.AddPoint(dp)
		}
	}
	return sliced
}

func Zip(timestamps []time.Time, values []float64) (TimeSeries, error) {
	if len(timestamps) != len(values) {
		return TimeSeries{}, errors.New("timestamps and values slices must have the same length")
	}

	points := make([]DataPoint, len(timestamps))
	for i := range timestamps {
		points[i] = DataPoint{
			timestamp: timestamps[i],
			value:     values[i],
		}
	}
	return TimeSeries{datapoints: points}, nil
}

func (ts *TimeSeries) UnZip() ([]time.Time, []float64) {
	timestamps := make([]time.Time, len(ts.datapoints))
	values := make([]float64, len(ts.datapoints))
	for i, point := range ts.datapoints {
		timestamps[i] = point.timestamp
		values[i] = point.value
	}
	return timestamps, values
}

/**
 * Maps a function over the values of the TimeSeries.
 *
 * @param f A function that takes a float64 and returns a float64.
 *
 * @return A new TimeSeries with the function applied to each value.
 */
func (ts *TimeSeries) MapValues(f func(float64) float64) TimeSeries {
	mapped := Empty()
	for _, dp := range ts.datapoints {
		mapped.AddPoint(DataPoint{
			timestamp: dp.timestamp,
			value:     f(dp.value),
		})
	}
	return mapped
}

/**
 * Filters the TimeSeries based on a predicate function.
 *
 * @param f A function that takes a DataPoint and returns a bool indicating whether to include the DataPoint.
 *
 * @return A new TimeSeries containing only the DataPoints that satisfy the predicate.
 */
func (ts *TimeSeries) Filter(f func(DataPoint) bool) TimeSeries {
	filtered := Empty()
	for _, dp := range ts.datapoints {
		if f(dp) {
			filtered.AddPoint(dp)
		}
	}
	return filtered
}

/**
 * Groups the TimeSeries by a specified time function and aggregates the values using a provided function.
 *
 * @param g A function that takes a time.Time and returns a grouped time.Time (e.g., rounding to the nearest hour).
 * @param f A function that takes a slice of DataPoint and returns a float64 representing the aggregated value (e.g., sum, average).
 * @return A new TimeSeries with grouped timestamps and aggregated values.
 */
func (ts *TimeSeries) GroupByTime(g func(dt time.Time) time.Time, f func(dp []DataPoint) float64) TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	} else {
		var grouped [][]DataPoint
		for _, dp := range ts.datapoints {
			groupedKey := g(dp.timestamp)
			idx, err := findIndexInGroup(grouped, groupedKey)
			if err == nil {
				grouped[idx] = append(grouped[idx], dp)
			} else {
				grouped = append(grouped, []DataPoint{dp})
			}
		}
		var result []DataPoint
		for _, group := range grouped {
			result = append(result, DataPoint{timestamp: g(group[0].timestamp), value: f(group)})

		}
		return TimeSeries{result}
	}
}

/**
 * Merges two TimeSeries into one, combining their DataPoints in chronological order.
 * If both TimeSeries have DataPoints with the same timestamp, the DataPoint from the first TimeSeries is retained.
 *
 * @param otherTS The other TimeSeries to merge with.
 *
 * @return A new TimeSeries containing all DataPoints from both TimeSeries in chronological order.
 */
func (ts *TimeSeries) Merge(otherTS TimeSeries) TimeSeries {
	merged := Empty()
	tsi, otsi := 0, 0
	for tsi < ts.Length() && otsi < otherTS.Length() {
		if ts.datapoints[tsi].timestamp.Before(otherTS.datapoints[otsi].timestamp) {
			merged.AddPoint(ts.datapoints[tsi])
			tsi++
		} else if ts.datapoints[tsi].timestamp.Equal(otherTS.datapoints[otsi].timestamp) {
			merged.AddPoint(ts.datapoints[tsi])
			tsi++
			otsi++
		} else {
			merged.AddPoint(otherTS.datapoints[otsi])
			otsi++
		}
	}

	for tsi < ts.Length() {
		merged.AddPoint(ts.datapoints[tsi])
		tsi++
	}

	for otsi < otherTS.Length() {
		merged.AddPoint(otherTS.datapoints[otsi])
		otsi++
	}

	return merged
}

/**
 * Joins
 */

/**
 * Joins (Inner) two TimeSeries on their timestamps.
 *
 * @param otherTS The other TimeSeries to join with.
 *
 * @return A AlignedSeries containing DataPoints with matching timestamps from both TimeSeries.
 */
func (ts *TimeSeries) Join(otherTS TimeSeries) AlignedSeries {
	if ts.IsEmpty() || otherTS.IsEmpty() {
		return AlignedSeries{}
	} else {
		res := AlignedSeries{}

		for _, leftValue := range ts.datapoints {
			for _, rightValue := range otherTS.datapoints {
				if leftValue.timestamp.Equal(rightValue.timestamp) {
					res.datapoints = append(res.datapoints, DoubleDataPoint{
						timestamp:  leftValue.timestamp,
						leftValue:  leftValue.value,
						rightValue: rightValue.value,
					})
				}
			}
		}
		return res
	}
}

/**
 * Joins (Left) two TimeSeries on their timestamps.
 *
 * @param otherTS The other TimeSeries to join with.
 * @param defaultValue The default value to use for missing right-side DataPoints.
 *
 * @return A AlignedSeries containing DataPoints from the left TimeSeries and matching DataPoints from the right TimeSeries, using defaultValue for missing matches.
 */
func (ts *TimeSeries) JoinLeft(otherTS TimeSeries, defaultValue float64) AlignedSeries {
	if ts.IsEmpty() {
		return AlignedSeries{}
	} else {
		res := AlignedSeries{}

		for _, leftValue := range ts.datapoints {
			matched := false
			for _, rightValue := range otherTS.datapoints {
				if leftValue.timestamp.Equal(rightValue.timestamp) {
					res.datapoints = append(res.datapoints, DoubleDataPoint{
						timestamp:  leftValue.timestamp,
						leftValue:  leftValue.value,
						rightValue: rightValue.value,
					})
					matched = true
					break
				}
			}
			if !matched {
				res.datapoints = append(res.datapoints, DoubleDataPoint{
					timestamp:  leftValue.timestamp,
					leftValue:  leftValue.value,
					rightValue: defaultValue,
				})
			}
		}
		return res
	}
}

func (ts *TimeSeries) JoinOuter(otherTS TimeSeries, defaultLeftValue float64, defaultRightValue float64) AlignedSeries {
	if ts.IsEmpty() && otherTS.IsEmpty() {
		return AlignedSeries{}
	} else {
		res := AlignedSeries{}
		for _, leftValue := range ts.datapoints {
			matched := false
			for _, rightValue := range otherTS.datapoints {
				if leftValue.timestamp.Equal(rightValue.timestamp) {
					res.datapoints = append(res.datapoints, DoubleDataPoint{
						timestamp:  leftValue.timestamp,
						leftValue:  leftValue.value,
						rightValue: rightValue.value,
					})
					matched = true
					break
				}
			}
			if !matched {
				res.datapoints = append(res.datapoints, DoubleDataPoint{
					timestamp:  leftValue.timestamp,
					leftValue:  leftValue.value,
					rightValue: defaultRightValue,
				})
			}
		}
		for _, rightValue := range otherTS.datapoints {
			matched := false
			for _, leftValue := range ts.datapoints {
				if rightValue.timestamp.Equal(leftValue.timestamp) {
					matched = true
					break
				}
			}
			if !matched {
				res.datapoints = append(res.datapoints, DoubleDataPoint{
					timestamp:  rightValue.timestamp,
					leftValue:  defaultLeftValue,
					rightValue: rightValue.value,
				})
			}
		}
		return res
	}
}

/**
* Statistics
 */

/**
 * Finds the minimum value in the TimeSeries.
 *
 * @return The DataPoint with the minimum value, or an error if the TimeSeries is empty.
 */
func (ts *TimeSeries) Min() (DataPoint, error) {
	if ts.IsEmpty() {
		return DataPoint{}, errors.New("timeseries is empty")
	}
	minDP := ts.datapoints[0]
	for _, dp := range ts.datapoints {
		if dp.value < minDP.value {
			minDP = dp
		}
	}
	return minDP, nil
}

/**
 * Calculates the sum of all values in the TimeSeries.
 *
 * @return The sum of the values. Returns 0.0 if the TimeSeries is empty.
 */
func (ts *TimeSeries) Sum() float64 {
	if ts.IsEmpty() {
		return 0.0
	}
	sum := 0.0
	for _, dp := range ts.datapoints {
		sum += dp.value
	}
	return sum
}

/**
 * Finds the maximum value in the TimeSeries.
 *
 * @return The DataPoint with the maximum value, or an error if the TimeSeries is empty.
 */
func (ts *TimeSeries) Max() (DataPoint, error) {
	if ts.IsEmpty() {
		return DataPoint{}, errors.New("timeseries is empty")
	}
	maxDP := ts.datapoints[0]
	for _, dp := range ts.datapoints {
		if dp.value > maxDP.value {
			maxDP = dp
		}
	}
	return maxDP, nil
}

func findIndexInGroup(grouped [][]DataPoint, key time.Time) (int, error) {
	for i, k := range grouped {
		if len(k) == 0 {
			return -1, errors.New("empty group encountered")
		}
		if k[0].timestamp.Equal(key) {
			return i, nil
		}
	}
	return -1, errors.New("key not found in groups")
}
