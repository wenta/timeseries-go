package timeseriesgo

import (
	"errors"
	"fmt"
	"math"
	"time"
)

type DataPoint struct {
	Timestamp time.Time
	Value     float64
}

type TimeSeries struct {
	datapoints []DataPoint
}

func Empty() TimeSeries {
	return TimeSeries{datapoints: []DataPoint{}}
}

// FromDataPoints builds a TimeSeries from a slice of datapoints (copied).
func FromDataPoints(points []DataPoint) TimeSeries {
	cp := make([]DataPoint, len(points))
	copy(cp, points)
	return TimeSeries{datapoints: cp}
}

func (ts *TimeSeries) IsEmpty() bool {
	return len(ts.datapoints) == 0
}

func (ts *TimeSeries) Length() int {
	return len(ts.datapoints)
}

/**
 * Returns the values of all points.
 */
func (ts *TimeSeries) Values() []float64 {
	var res []float64
	for _, dp := range ts.datapoints {
		res = append(res, dp.Value)
	}
	return res
}

/**
 * Returns all timestamps.
 */
func (ts *TimeSeries) Timestamps() []time.Time {
	var res []time.Time
	for _, dp := range ts.datapoints {
		res = append(res, dp.Timestamp)
	}
	return res
}

// DataPoints returns a shallow copy of underlying datapoints to allow safe read access.
func (ts *TimeSeries) DataPoints() []DataPoint {
	cp := make([]DataPoint, len(ts.datapoints))
	copy(cp, ts.datapoints)
	return cp
}

/**
 * Returns the last point in the series.
 */
func (ts *TimeSeries) Last() (DataPoint, error) {
	if ts.IsEmpty() {
		return DataPoint{}, errors.New("timeSeries is empty")
	}
	return ts.datapoints[len(ts.datapoints)-1], nil
}

/**
 * Returns the first point in the series.
 */
func (ts *TimeSeries) Head() (DataPoint, error) {
	if ts.IsEmpty() {
		return DataPoint{}, errors.New("timeSeries is empty")
	}
	return ts.datapoints[0], nil
}

/**
 * Returns the series without the first point.
 */
func (ts *TimeSeries) Tail() TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}
	cloned := make([]DataPoint, len(ts.datapoints)-1)
	copy(cloned, ts.datapoints[1:])
	return TimeSeries{datapoints: cloned}
}

/**
 * Most frequent interval between consecutive points.
 */
func (ts *TimeSeries) Resolution() (time.Duration, error) {
	if ts.IsEmpty() {
		return 0 * time.Second, errors.New("timeSeries is empty")
	} else if ts.Length() == 1 {
		return 0 * time.Second, errors.New("timeSeries has just one point")
	}

	var modeDuration time.Duration
	var modeCount int
	counts := make(map[time.Duration]int)

	for i := 1; i < len(ts.datapoints); i++ {
		d := ts.datapoints[i].Timestamp.Sub(ts.datapoints[i-1].Timestamp)
		counts[d]++
	}

	for d, c := range counts {
		if c > modeCount || (c == modeCount && (modeCount == 0 || d < modeDuration)) {
			modeCount = c
			modeDuration = d
		}
	}

	return modeDuration, nil
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
		fmt.Printf("%s, %.2f\n", dp.Timestamp.Format(time.RFC3339), dp.Value)
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
		if (dp.Timestamp.Equal(start) || dp.Timestamp.After(start)) && dp.Timestamp.Before(end) {
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
			Timestamp: timestamps[i],
			Value:     values[i],
		}
	}
	return TimeSeries{datapoints: points}, nil
}

/**
 * Splits the series into separate slices of timestamps and values.
 */
func (ts *TimeSeries) UnZip() ([]time.Time, []float64) {
	timestamps := make([]time.Time, len(ts.datapoints))
	values := make([]float64, len(ts.datapoints))
	for i, point := range ts.datapoints {
		timestamps[i] = point.Timestamp
		values[i] = point.Value
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
			Timestamp: dp.Timestamp,
			Value:     f(dp.Value),
		})
	}
	return mapped
}

/**
 * Maps over the full DataPoint.
 */
func (ts *TimeSeries) Map(f func(DataPoint) DataPoint) TimeSeries {
	mapped := Empty()
	for _, dp := range ts.datapoints {
		mapped.AddPoint(f(dp))
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
			groupedKey := g(dp.Timestamp)
			idx, err := findIndexInGroup(grouped, groupedKey)
			if err == nil {
				grouped[idx] = append(grouped[idx], dp)
			} else {
				grouped = append(grouped, []DataPoint{dp})
			}
		}
		var result []DataPoint
		for _, group := range grouped {
			result = append(result, DataPoint{Timestamp: g(group[0].Timestamp), Value: f(group)})

		}
		return TimeSeries{result}
	}
}

func (ts TimeSeries) RollingWindow(window time.Duration, f func(vs []float64) float64) TimeSeries {
	return ts.Map(func(dp DataPoint) DataPoint {
		ws := ts.Filter(func(dp2 DataPoint) bool {
			return (dp2.Timestamp.Before(dp.Timestamp) && dp2.Timestamp.After(dp.Timestamp.Add(-window))) || dp2.Timestamp.Equal(dp.Timestamp)
		})
		v := f(ws.Values())
		return DataPoint{Timestamp: dp.Timestamp, Value: v}
	})
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
		if ts.datapoints[tsi].Timestamp.Before(otherTS.datapoints[otsi].Timestamp) {
			merged.AddPoint(ts.datapoints[tsi])
			tsi++
		} else if ts.datapoints[tsi].Timestamp.Equal(otherTS.datapoints[otsi].Timestamp) {
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
 * Joins (inner) two TimeSeries on their timestamps.
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
				if leftValue.Timestamp.Equal(rightValue.Timestamp) {
					res.datapoints = append(res.datapoints, DoubleDataPoint{
						Timestamp:  leftValue.Timestamp,
						LeftValue:  leftValue.Value,
						RightValue: rightValue.Value,
					})
				}
			}
		}
		return res
	}
}

/**
 * Joins (left) two TimeSeries on their timestamps, filling missing right values with a default.
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
				if leftValue.Timestamp.Equal(rightValue.Timestamp) {
					res.datapoints = append(res.datapoints, DoubleDataPoint{
						Timestamp:  leftValue.Timestamp,
						LeftValue:  leftValue.Value,
						RightValue: rightValue.Value,
					})
					matched = true
					break
				}
			}
			if !matched {
				res.datapoints = append(res.datapoints, DoubleDataPoint{
					Timestamp:  leftValue.Timestamp,
					LeftValue:  leftValue.Value,
					RightValue: defaultValue,
				})
			}
		}
		return res
	}
}

/**
 * Joins (outer) two TimeSeries on their timestamps, filling missing values with defaults.
 */
func (ts *TimeSeries) JoinOuter(otherTS TimeSeries, defaultLeftValue float64, defaultRightValue float64) AlignedSeries {
	if ts.IsEmpty() && otherTS.IsEmpty() {
		return AlignedSeries{}
	} else {
		res := AlignedSeries{}
		for _, leftValue := range ts.datapoints {
			matched := false
			for _, rightValue := range otherTS.datapoints {
				if leftValue.Timestamp.Equal(rightValue.Timestamp) {
					res.datapoints = append(res.datapoints, DoubleDataPoint{
						Timestamp:  leftValue.Timestamp,
						LeftValue:  leftValue.Value,
						RightValue: rightValue.Value,
					})
					matched = true
					break
				}
			}
			if !matched {
				res.datapoints = append(res.datapoints, DoubleDataPoint{
					Timestamp:  leftValue.Timestamp,
					LeftValue:  leftValue.Value,
					RightValue: defaultRightValue,
				})
			}
		}
		for _, rightValue := range otherTS.datapoints {
			matched := false
			for _, leftValue := range ts.datapoints {
				if rightValue.Timestamp.Equal(leftValue.Timestamp) {
					matched = true
					break
				}
			}
			if !matched {
				res.datapoints = append(res.datapoints, DoubleDataPoint{
					Timestamp:  rightValue.Timestamp,
					LeftValue:  defaultLeftValue,
					RightValue: rightValue.Value,
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
		if dp.Value < minDP.Value {
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
		sum += dp.Value
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
		if dp.Value > maxDP.Value {
			maxDP = dp
		}
	}
	return maxDP, nil
}

func (ts *TimeSeries) Percentile(p int) (float64, error) {
	if ts.IsEmpty() {
		return 0.0, errors.New("timeseries is empty")
	}
	vs := ts.Values()
	vsLen := len(vs)
	pos := float64(p*(vsLen+1)) / 100
	if pos < 1 {
		return vs[0], nil
	} else if pos >= float64(vsLen) {
		return vs[vsLen-1], nil
	} else {
		pf := int(math.Floor(pos))
		lower := vs[pf-1]
		upper := vs[pf]
		d := pos - math.Floor(pos)
		p := lower + d*(upper-lower)
		return p, nil
	}
}

func (ts *TimeSeries) Median() (float64, error) {
	return ts.Percentile(50)
}

func findIndexInGroup(grouped [][]DataPoint, key time.Time) (int, error) {
	for i, k := range grouped {
		if len(k) == 0 {
			return -1, errors.New("empty group encountered")
		}
		if k[0].Timestamp.Equal(key) {
			return i, nil
		}
	}
	return -1, errors.New("key not found in groups")
}
