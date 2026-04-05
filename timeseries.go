package timeseriesgo

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"time"
)

type DataPoint struct {
	Timestamp time.Time
	Value     float64
}

type TimeSeries struct {
	datapoints []DataPoint
	label      string
}

/**
 * Creates an empty TimeSeries with a default label.
 *
 * @return A new empty TimeSeries.
 */
func Empty() TimeSeries {
	return TimeSeries{datapoints: []DataPoint{}, label: "new series"}
}

/**
 * Creates an empty TimeSeries with the provided label.
 *
 * @param label The label to assign to the created TimeSeries.
 *
 * @return A new empty TimeSeries with the provided label.
 */
func EmptyLabeled(label string) TimeSeries {
	return TimeSeries{datapoints: []DataPoint{}, label: label}
}

/**
 * Builds a TimeSeries from a slice of datapoints.
 *
 * @param points The datapoints to copy into the new TimeSeries.
 *
 * @return A new TimeSeries containing a copy of the provided datapoints.
 */
func FromDataPoints(points []DataPoint) TimeSeries {
	cp := make([]DataPoint, len(points))
	copy(cp, points)
	return TimeSeries{datapoints: cp}
}

/**
 * Checks whether the TimeSeries is empty.
 *
 * @return True if the series contains no datapoints, otherwise false.
 */
func (ts *TimeSeries) IsEmpty() bool {
	return len(ts.datapoints) == 0
}

/**
 * Returns the number of datapoints in the TimeSeries.
 *
 * @return The number of datapoints in the series.
 */
func (ts *TimeSeries) Length() int {
	return len(ts.datapoints)
}

/**
 * Returns the values of all points.
 *
 * @return A slice of float64 values in series order.
 */
func (ts *TimeSeries) Values() []float64 {
	res := make([]float64, len(ts.datapoints))
	for i, dp := range ts.datapoints {
		res[i] = dp.Value
	}
	return res
}

/**
 * Returns all timestamps.
 *
 * @return A slice of timestamps in series order.
 */
func (ts *TimeSeries) Timestamps() []time.Time {
	res := make([]time.Time, len(ts.datapoints))
	for i, dp := range ts.datapoints {
		res[i] = dp.Timestamp
	}
	return res
}

/**
 * Returns the datapoints of the TimeSeries.
 *
 * @return A shallow copy of the underlying DataPoint slice for safe read-only access.
 */
func (ts *TimeSeries) DataPoints() []DataPoint {
	cp := make([]DataPoint, len(ts.datapoints))
	copy(cp, ts.datapoints)
	return cp
}

/**
 * Returns the last point in the series.
 *
 * @return The last DataPoint in the series, or an error if the series is empty.
 */
func (ts *TimeSeries) Last() (DataPoint, error) {
	if ts.IsEmpty() {
		return DataPoint{}, errors.New("timeSeries is empty")
	}
	return ts.datapoints[len(ts.datapoints)-1], nil
}

/**
 * Returns the first point in the series.
 *
 * @return The first DataPoint in the series, or an error if the series is empty.
 */
func (ts *TimeSeries) Head() (DataPoint, error) {
	if ts.IsEmpty() {
		return DataPoint{}, errors.New("timeSeries is empty")
	}
	return ts.datapoints[0], nil
}

/**
 * Returns the series without the first point.
 *
 * @return A new TimeSeries without the first datapoint.
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
 *
 * @return The most common time interval between consecutive datapoints, or an error if it cannot be determined.
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
 *
 * @return None. The function writes the series to standard output.
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

/**
 * Combines timestamps and values into a TimeSeries.
 *
 * @param timestamps A slice of timestamps for the resulting datapoints.
 * @param values A slice of values for the resulting datapoints.
 *
 * @return A new TimeSeries built from the provided timestamps and values, or an error if the slice lengths differ.
 */
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
 *
 * @return A slice of timestamps and a slice of values extracted from the series.
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
 *
 * @param f A function that takes a DataPoint and returns a transformed DataPoint.
 *
 * @return A new TimeSeries with the function applied to each DataPoint.
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
 * Resamples the TimeSeries on a fixed interval.
 *
 * @param delta The step between consecutive timestamps in the resampled series.
 * @param f A function that takes the previous point, the next point, and the target timestamp, returning the interpolated value.
 *
 * @return A new TimeSeries on a regular grid from first to last timestamp (inclusive if last aligns to the grid).
 *         For each grid point: if an original exists, it is used; otherwise f(prev, next, target) provides the value.
 *         If delta <= 0, returns a copy. Empty input returns empty.
 */
func (ts *TimeSeries) Resample(delta time.Duration, f func(DataPoint, DataPoint, time.Time) float64) TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}
	if delta <= 0 {
		return FromDataPoints(ts.DataPoints())
	}

	points := ts.datapoints
	result := EmptyLabeled(ts.label + " resampled")
	result.datapoints = make([]DataPoint, 0, len(points))

	start := points[0].Timestamp
	end := points[len(points)-1].Timestamp
	prevIdx := -1
	nextIdx := 0

	for t := start; !t.After(end); t = t.Add(delta) {
		for nextIdx < len(points) && points[nextIdx].Timestamp.Before(t) {
			prevIdx = nextIdx
			nextIdx++
		}

		if nextIdx < len(points) && points[nextIdx].Timestamp.Equal(t) {
			result.AddPoint(points[nextIdx])
			prevIdx = nextIdx
			nextIdx++
			continue
		}

		if prevIdx >= 0 && nextIdx < len(points) {
			val := f(points[prevIdx], points[nextIdx], t)
			result.AddPoint(DataPoint{Timestamp: t, Value: val})
		}
	}

	return result
}

/**
 * Resamples the TimeSeries on a fixed interval, filling gaps with a default value.
 *
 * @param delta The step between consecutive timestamps in the resampled series.
 * @param defaultValue Value used for any interpolated points.
 *
 * @return A new TimeSeries containing all original points plus default-filled points. If delta <= 0, returns a copy. Empty input returns empty.
 */
func (ts *TimeSeries) ResampleWithDefaultValue(delta time.Duration, defaultValue float64) TimeSeries {
	return ts.Resample(delta, func(d1 DataPoint, d2 DataPoint, idx time.Time) float64 {
		return defaultValue
	})
}

/**
 * Interpolates the TimeSeries on a fixed interval using linear interpolation.
 *
 * @param delta The step between consecutive timestamps in the interpolated series.
 *
 * @return A new TimeSeries on a regular grid, with missing values linearly interpolated. If delta <= 0, returns a copy. Empty input returns empty.
 */
func (ts *TimeSeries) Interpolate(delta time.Duration) TimeSeries {
	return ts.Resample(delta, func(d1 DataPoint, d2 DataPoint, idx time.Time) float64 {
		total := d2.Timestamp.Sub(d1.Timestamp).Seconds()
		if total == 0 {
			return d1.Value
		}
		elapsed := idx.Sub(d1.Timestamp).Seconds()
		return d1.Value + (d2.Value-d1.Value)*(elapsed/total)
	})
}

/**
 * Steps the TimeSeries on a fixed interval, splitting each value evenly across the gap to the next point.
 * Example: value 10 at 1h, delta 15m -> four points of 2.5 within that hour.
 *
 * @param delta The step between consecutive timestamps in the stepped series.
 *
 * @return A new TimeSeries on a regular grid where each original value is distributed across sub-intervals.
 *         If delta <= 0, returns a copy. Empty input returns empty.
 */
func (ts *TimeSeries) Step(delta time.Duration) TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}
	if delta <= 0 {
		return FromDataPoints(ts.DataPoints())
	}

	points := ts.DataPoints()
	result := EmptyLabeled(ts.label + " step")

	for i := 0; i < len(points)-1; i++ {
		prev := points[i]
		next := points[i+1]

		gap := next.Timestamp.Sub(prev.Timestamp)
		if gap <= 0 {
			continue
		}
		steps := int(gap / delta)
		if steps == 0 {
			continue
		}

		fraction := prev.Value * delta.Seconds() / gap.Seconds()
		for j := 1; j <= steps; j++ {
			tsAt := prev.Timestamp.Add(time.Duration(j) * delta)
			result.AddPoint(DataPoint{Timestamp: tsAt, Value: fraction})
		}
	}

	return result
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
	}

	grouped := make(map[int64][]DataPoint)
	order := make([]time.Time, 0)
	for _, dp := range ts.datapoints {
		groupedKey := g(dp.Timestamp)
		key := groupedKey.UnixNano()
		if _, exists := grouped[key]; !exists {
			order = append(order, groupedKey)
		}
		grouped[key] = append(grouped[key], dp)
	}

	result := make([]DataPoint, 0, len(order))
	for _, groupedKey := range order {
		group := grouped[groupedKey.UnixNano()]
		result = append(result, DataPoint{Timestamp: groupedKey, Value: f(group)})
	}

	return TimeSeries{result, ts.label + " grouped"}
}

/**
 * Applies a rolling window aggregation over the TimeSeries.
 *
 * @param window The trailing time window used for each aggregation.
 * @param f A function that takes the values in the window and returns an aggregated float64.
 *
 * @return A new TimeSeries containing one aggregated value per original datapoint.
 */
func (ts TimeSeries) RollingWindow(window time.Duration, f func(vs []float64) float64) TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}

	result := EmptyLabeled(ts.label + " rolling")
	result.datapoints = make([]DataPoint, 0, len(ts.datapoints))

	left := 0
	windowValues := make([]float64, 0, len(ts.datapoints))
	for right, dp := range ts.datapoints {
		cutoff := dp.Timestamp.Add(-window)
		for left <= right && !ts.datapoints[left].Timestamp.After(cutoff) && !ts.datapoints[left].Timestamp.Equal(dp.Timestamp) {
			left++
		}

		windowValues = windowValues[:0]
		for i := left; i <= right; i++ {
			windowValues = append(windowValues, ts.datapoints[i].Value)
		}

		result.AddPoint(DataPoint{
			Timestamp: dp.Timestamp,
			Value:     f(windowValues),
		})
	}

	return result
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
		return EmptyLabeledAlignedSeries("empty series")
	}

	res := EmptyLabeledAlignedSeries(ts.label + " joined with " + otherTS.label)
	capacity := ts.Length()
	if otherTS.Length() < capacity {
		capacity = otherTS.Length()
	}
	res.datapoints = make([]DoubleDataPoint, 0, capacity)

	leftIdx, rightIdx := 0, 0
	for leftIdx < len(ts.datapoints) && rightIdx < len(otherTS.datapoints) {
		leftValue := ts.datapoints[leftIdx]
		rightValue := otherTS.datapoints[rightIdx]

		if leftValue.Timestamp.Before(rightValue.Timestamp) {
			leftIdx++
			continue
		}
		if rightValue.Timestamp.Before(leftValue.Timestamp) {
			rightIdx++
			continue
		}

		res.datapoints = append(res.datapoints, DoubleDataPoint{
			Timestamp:  leftValue.Timestamp,
			LeftValue:  leftValue.Value,
			RightValue: rightValue.Value,
		})
		leftIdx++
		rightIdx++
	}

	return res
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
		return EmptyLabeledAlignedSeries("empty series")
	}

	res := EmptyLabeledAlignedSeries(ts.label + " joined with " + otherTS.label)
	res.datapoints = make([]DoubleDataPoint, 0, ts.Length())

	rightIdx := 0
	for _, leftValue := range ts.datapoints {
		for rightIdx < len(otherTS.datapoints) && otherTS.datapoints[rightIdx].Timestamp.Before(leftValue.Timestamp) {
			rightIdx++
		}

		rightValue := defaultValue
		if rightIdx < len(otherTS.datapoints) && otherTS.datapoints[rightIdx].Timestamp.Equal(leftValue.Timestamp) {
			rightValue = otherTS.datapoints[rightIdx].Value
		}

		res.datapoints = append(res.datapoints, DoubleDataPoint{
			Timestamp:  leftValue.Timestamp,
			LeftValue:  leftValue.Value,
			RightValue: rightValue,
		})
	}

	return res
}

/**
 * Joins (outer) two TimeSeries on their timestamps, filling missing values with defaults.
 *
 * @param otherTS The other TimeSeries to join with.
 * @param defaultLeftValue The default value to use for missing left-side datapoints.
 * @param defaultRightValue The default value to use for missing right-side datapoints.
 *
 * @return An AlignedSeries containing datapoints from both TimeSeries, using default values for missing matches.
 */
func (ts *TimeSeries) JoinOuter(otherTS TimeSeries, defaultLeftValue float64, defaultRightValue float64) AlignedSeries {
	if ts.IsEmpty() && otherTS.IsEmpty() {
		return EmptyLabeledAlignedSeries("empty series")
	}

	res := EmptyLabeledAlignedSeries(ts.label + " joined with " + otherTS.label)
	res.datapoints = make([]DoubleDataPoint, 0, ts.Length()+otherTS.Length())

	leftIdx, rightIdx := 0, 0
	for leftIdx < len(ts.datapoints) && rightIdx < len(otherTS.datapoints) {
		leftValue := ts.datapoints[leftIdx]
		rightValue := otherTS.datapoints[rightIdx]

		if leftValue.Timestamp.Before(rightValue.Timestamp) {
			res.datapoints = append(res.datapoints, DoubleDataPoint{
				Timestamp:  leftValue.Timestamp,
				LeftValue:  leftValue.Value,
				RightValue: defaultRightValue,
			})
			leftIdx++
			continue
		}

		if rightValue.Timestamp.Before(leftValue.Timestamp) {
			res.datapoints = append(res.datapoints, DoubleDataPoint{
				Timestamp:  rightValue.Timestamp,
				LeftValue:  defaultLeftValue,
				RightValue: rightValue.Value,
			})
			rightIdx++
			continue
		}

		res.datapoints = append(res.datapoints, DoubleDataPoint{
			Timestamp:  leftValue.Timestamp,
			LeftValue:  leftValue.Value,
			RightValue: rightValue.Value,
		})
		leftIdx++
		rightIdx++
	}

	for leftIdx < len(ts.datapoints) {
		leftValue := ts.datapoints[leftIdx]
		res.datapoints = append(res.datapoints, DoubleDataPoint{
			Timestamp:  leftValue.Timestamp,
			LeftValue:  leftValue.Value,
			RightValue: defaultRightValue,
		})
		leftIdx++
	}

	for rightIdx < len(otherTS.datapoints) {
		rightValue := otherTS.datapoints[rightIdx]
		res.datapoints = append(res.datapoints, DoubleDataPoint{
			Timestamp:  rightValue.Timestamp,
			LeftValue:  defaultLeftValue,
			RightValue: rightValue.Value,
		})
		rightIdx++
	}

	return res
}

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

/**
 * Calculates the percentile value of the TimeSeries.
 *
 * @param p The percentile to calculate, expressed as an integer from 0 to 100.
 *
 * @return The percentile value, or an error if the series is empty.
 */
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

/**
 * Returns a new TimeSeries containing differences between consecutive points.
 *
 * @return A new TimeSeries with one datapoint per consecutive difference.
 */
func (ts *TimeSeries) Differentiate() TimeSeries {
	if ts.Length() < 2 {
		return Empty()
	}

	result := EmptyLabeled(ts.label + " differentiated")
	result.datapoints = make([]DataPoint, 0, len(ts.datapoints)-1)
	prev := ts.datapoints[0]
	for i := 1; i < len(ts.datapoints); i++ {
		dp := ts.datapoints[i]
		result.AddPoint(DataPoint{Timestamp: dp.Timestamp, Value: dp.Value - prev.Value})
		prev = dp
	}

	return result
}

/**
 * Returns a new TimeSeries containing pairwise sums of consecutive points.
 *
 * @return A new TimeSeries with one datapoint per pairwise sum.
 */
func (ts *TimeSeries) Integrate() TimeSeries {
	if ts.Length() < 2 {
		return Empty()
	}

	result := EmptyLabeled(ts.label + " integrated")
	result.datapoints = make([]DataPoint, 0, len(ts.datapoints)-1)
	prev := ts.datapoints[0]
	for i := 1; i < len(ts.datapoints); i++ {
		dp := ts.datapoints[i]
		result.AddPoint(DataPoint{Timestamp: dp.Timestamp, Value: dp.Value + prev.Value})
		prev = dp
	}

	return result
}

/**
 * Calculates the median value of the TimeSeries.
 *
 * @return The median value, or an error if the series is empty.
 */
func (ts *TimeSeries) Median() (float64, error) {
	return ts.Percentile(50)
}

/**
 * Returns a lagged TimeSeries shifted backward by a fixed number of points.
 *
 * Lag(1) places the previous value at the current timestamp. The first
 * `steps` timestamps are omitted because no lagged value exists for them.
 *
 * @param steps The number of points to lag the values by.
 *
 * @return A new TimeSeries with lagged values. If steps <= 0, returns a copy of the original series. If steps >= series length, returns an empty series.
 */
func (ts *TimeSeries) Lag(steps int) TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}
	if steps <= 0 {
		return cloneSeries(ts)
	}
	if steps >= len(ts.datapoints) {
		return EmptyLabeled(ts.label + " lagged")
	}

	result := make([]DataPoint, len(ts.datapoints)-steps)
	for i := steps; i < len(ts.datapoints); i++ {
		result[i-steps] = DataPoint{
			Timestamp: ts.datapoints[i].Timestamp,
			Value:     ts.datapoints[i-steps].Value,
		}
	}
	return TimeSeries{datapoints: result, label: ts.label + " lagged"}
}

/**
 * Returns a TimeSeries with all timestamps shifted by the provided duration.
 *
 * Values remain unchanged and relative spacing between datapoints is preserved.
 *
 * @param delta The duration added to every timestamp. Negative values shift the series backward in time.
 *
 * @return A new TimeSeries with shifted timestamps.
 */
func (ts *TimeSeries) Shift(delta time.Duration) TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}

	result := make([]DataPoint, len(ts.datapoints))
	for i, dp := range ts.datapoints {
		result[i] = DataPoint{
			Timestamp: dp.Timestamp.Add(delta),
			Value:     dp.Value,
		}
	}
	return TimeSeries{datapoints: result, label: ts.label + " shifted"}
}

/**
 * Returns the cumulative sum of the TimeSeries values.
 *
 * Each output point equals the sum of all values up to and including the
 * current point. Timestamps are preserved.
 *
 * @return A new TimeSeries containing cumulative sums.
 */
func (ts *TimeSeries) CumulativeSum() TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}

	result := make([]DataPoint, len(ts.datapoints))
	running := 0.0
	for i, dp := range ts.datapoints {
		running += dp.Value
		result[i] = DataPoint{
			Timestamp: dp.Timestamp,
			Value:     running,
		}
	}
	return TimeSeries{datapoints: result, label: ts.label + " cumulative sum"}
}

/**
 * Returns the cumulative minimum of the TimeSeries values.
 *
 * Each output point equals the minimum value observed so far. Timestamps are preserved.
 *
 * @return A new TimeSeries containing cumulative minima.
 */
func (ts *TimeSeries) CumulativeMin() TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}

	result := make([]DataPoint, len(ts.datapoints))
	currentMin := ts.datapoints[0].Value
	for i, dp := range ts.datapoints {
		if dp.Value < currentMin {
			currentMin = dp.Value
		}
		result[i] = DataPoint{
			Timestamp: dp.Timestamp,
			Value:     currentMin,
		}
	}
	return TimeSeries{datapoints: result, label: ts.label + " cumulative min"}
}

/**
 * Returns the cumulative maximum of the TimeSeries values.
 *
 * Each output point equals the maximum value observed so far. Timestamps are preserved.
 *
 * @return A new TimeSeries containing cumulative maxima.
 */
func (ts *TimeSeries) CumulativeMax() TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}

	result := make([]DataPoint, len(ts.datapoints))
	currentMax := ts.datapoints[0].Value
	for i, dp := range ts.datapoints {
		if dp.Value > currentMax {
			currentMax = dp.Value
		}
		result[i] = DataPoint{
			Timestamp: dp.Timestamp,
			Value:     currentMax,
		}
	}
	return TimeSeries{datapoints: result, label: ts.label + " cumulative max"}
}

/**
 * Forward-fills the TimeSeries onto a regular timestamp grid.
 *
 * The output grid starts at the first timestamp and advances by `delta` until
 * it reaches or passes the last timestamp. Original values are preserved on
 * exact grid matches, and missing grid points use the most recent previous value.
 *
 * @param delta The spacing of the regular output grid.
 *
 * @return A new regularized TimeSeries using forward fill. If delta <= 0, returns a copy. Empty input returns empty.
 */
func (ts *TimeSeries) ForwardFill(delta time.Duration) TimeSeries {
	return ts.fillRegular(delta, fillForward, 0)
}

/**
 * Backward-fills the TimeSeries onto a regular timestamp grid.
 *
 * The output grid starts at the first timestamp and advances by `delta` until
 * it reaches or passes the last timestamp. Original values are preserved on
 * exact grid matches, and missing grid points use the next available value.
 *
 * @param delta The spacing of the regular output grid.
 *
 * @return A new regularized TimeSeries using backward fill. If delta <= 0, returns a copy. Empty input returns empty.
 */
func (ts *TimeSeries) BackwardFill(delta time.Duration) TimeSeries {
	return ts.fillRegular(delta, fillBackward, 0)
}

/**
 * Fills missing points on a regular timestamp grid using a default value.
 *
 * The output grid starts at the first timestamp and advances by `delta` until
 * it reaches or passes the last timestamp. Original values are preserved on
 * exact grid matches, while missing grid points are assigned `defaultValue`.
 *
 * @param delta The spacing of the regular output grid.
 * @param defaultValue The value used for grid points that do not exist in the original series.
 *
 * @return A new regularized TimeSeries. If delta <= 0, returns a copy. Empty input returns empty.
 */
func (ts *TimeSeries) FillMissing(delta time.Duration, defaultValue float64) TimeSeries {
	return ts.fillRegular(delta, fillConstant, defaultValue)
}

/**
 * Checks whether the TimeSeries timestamps are sorted in non-decreasing order.
 *
 * Duplicate timestamps are allowed and still considered sorted.
 *
 * @return True if the timestamps are sorted in non-decreasing order, otherwise false.
 */
func (ts *TimeSeries) IsSorted() bool {
	for i := 1; i < len(ts.datapoints); i++ {
		if ts.datapoints[i].Timestamp.Before(ts.datapoints[i-1].Timestamp) {
			return false
		}
	}
	return true
}

/**
 * Returns a copy of the TimeSeries sorted by timestamp.
 *
 * The sort is stable, so datapoints with equal timestamps preserve their original relative order.
 *
 * @return A new TimeSeries sorted by timestamp.
 */
func (ts *TimeSeries) SortByTimestamp() TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}

	sorted := ts.DataPoints()
	slices.SortStableFunc(sorted, func(left DataPoint, right DataPoint) int {
		return left.Timestamp.Compare(right.Timestamp)
	})
	return TimeSeries{datapoints: sorted, label: ts.label + " sorted"}
}

/**
 * Checks whether the TimeSeries contains duplicate timestamps.
 *
 * @return True if at least two datapoints share the same timestamp, otherwise false.
 */
func (ts *TimeSeries) HasDuplicateTimestamps() bool {
	if len(ts.datapoints) < 2 {
		return false
	}

	seen := make(map[int64]struct{}, len(ts.datapoints))
	for _, dp := range ts.datapoints {
		key := dp.Timestamp.UnixNano()
		if _, exists := seen[key]; exists {
			return true
		}
		seen[key] = struct{}{}
	}
	return false
}

/**
 * Checks whether the TimeSeries uses a regular positive interval between consecutive timestamps.
 *
 * Empty and single-point series are treated as regular.
 *
 * @return True if the series is sorted and all consecutive timestamp gaps are equal and positive, otherwise false.
 */
func (ts *TimeSeries) IsRegular() bool {
	if len(ts.datapoints) < 2 {
		return true
	}

	step := ts.datapoints[1].Timestamp.Sub(ts.datapoints[0].Timestamp)
	if step <= 0 {
		return false
	}

	for i := 2; i < len(ts.datapoints); i++ {
		if ts.datapoints[i].Timestamp.Sub(ts.datapoints[i-1].Timestamp) != step {
			return false
		}
	}
	return true
}

/**
 * Reindexes the TimeSeries onto an explicit target index using exact timestamp matches.
 *
 * For duplicate timestamps in the source series, the last matching datapoint is used.
 * The output preserves the order of the provided index and always has the same length as `index`.
 *
 * @param index The target timestamps for the reindexed series.
 * @param defaultValue The value used for timestamps that are missing in the source series.
 *
 * @return A new TimeSeries aligned to the provided index.
 */
func (ts *TimeSeries) Reindex(index []time.Time, defaultValue float64) TimeSeries {
	if len(index) == 0 {
		return Empty()
	}

	valueByTimestamp := make(map[int64]float64, len(ts.datapoints))
	for _, dp := range ts.datapoints {
		valueByTimestamp[dp.Timestamp.UnixNano()] = dp.Value
	}

	result := make([]DataPoint, len(index))
	for i, timestamp := range index {
		value := defaultValue
		if exact, exists := valueByTimestamp[timestamp.UnixNano()]; exists {
			value = exact
		}
		result[i] = DataPoint{
			Timestamp: timestamp,
			Value:     value,
		}
	}
	return TimeSeries{datapoints: result, label: ts.label + " reindexed"}
}

/**
 * Reindexes the TimeSeries using the nearest timestamp within a tolerance.
 *
 * Only target timestamps whose nearest source timestamp is within `tolerance`
 * are included in the result. The output preserves the order of the provided index.
 * For duplicate timestamps in the source series, the last matching datapoint is used.
 *
 * @param index The target timestamps to align to.
 * @param tolerance The maximum allowed absolute timestamp difference for a match.
 *
 * @return A new TimeSeries containing only target timestamps that found a nearest match within tolerance. Negative tolerance returns an empty series.
 */
func (ts *TimeSeries) ReindexNearest(index []time.Time, tolerance time.Duration) TimeSeries {
	if len(index) == 0 || tolerance < 0 || ts.IsEmpty() {
		return Empty()
	}

	points := deduplicatedSortedPoints(ts.datapoints)
	timestamps := make([]time.Time, len(points))
	for i, point := range points {
		timestamps[i] = point.Timestamp
	}

	result := make([]DataPoint, 0, len(index))
	for _, target := range index {
		candidate, ok := nearestPointWithinTolerance(points, timestamps, target, tolerance)
		if !ok {
			continue
		}
		result = append(result, DataPoint{
			Timestamp: target,
			Value:     candidate.Value,
		})
	}
	return TimeSeries{datapoints: result, label: ts.label + " reindexed nearest"}
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

type fillMode int

const (
	fillForward fillMode = iota
	fillBackward
	fillConstant
)

func cloneSeries(ts *TimeSeries) TimeSeries {
	return TimeSeries{
		datapoints: ts.DataPoints(),
		label:      ts.label,
	}
}

func (ts *TimeSeries) fillRegular(delta time.Duration, mode fillMode, defaultValue float64) TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}
	if delta <= 0 {
		return cloneSeries(ts)
	}

	points := deduplicatedSortedPoints(ts.datapoints)
	if len(points) == 0 {
		return Empty()
	}

	targets := regularGrid(points[0].Timestamp, points[len(points)-1].Timestamp, delta)
	if len(targets) == 0 {
		return Empty()
	}

	switch mode {
	case fillBackward:
		return backwardFillSeries(points, targets, ts.label+" backward filled")
	case fillConstant:
		return constantFillSeries(points, targets, defaultValue, ts.label+" filled")
	default:
		return forwardFillSeries(points, targets, ts.label+" forward filled")
	}
}

func forwardFillSeries(points []DataPoint, targets []time.Time, label string) TimeSeries {
	result := make([]DataPoint, 0, len(targets))
	sourceIdx := 0
	hasLast := false
	lastValue := 0.0

	for _, target := range targets {
		for sourceIdx < len(points) && points[sourceIdx].Timestamp.Before(target) {
			lastValue = points[sourceIdx].Value
			hasLast = true
			sourceIdx++
		}

		if sourceIdx < len(points) && points[sourceIdx].Timestamp.Equal(target) {
			lastValue = points[sourceIdx].Value
			hasLast = true
			result = append(result, DataPoint{Timestamp: target, Value: lastValue})
			sourceIdx++
			continue
		}

		if hasLast {
			result = append(result, DataPoint{Timestamp: target, Value: lastValue})
		}
	}

	return TimeSeries{datapoints: result, label: label}
}

func backwardFillSeries(points []DataPoint, targets []time.Time, label string) TimeSeries {
	result := make([]DataPoint, len(targets))
	sourceIdx := len(points) - 1
	hasNext := false
	nextValue := 0.0

	for i := len(targets) - 1; i >= 0; i-- {
		target := targets[i]
		for sourceIdx >= 0 && points[sourceIdx].Timestamp.After(target) {
			nextValue = points[sourceIdx].Value
			hasNext = true
			sourceIdx--
		}

		if sourceIdx >= 0 && points[sourceIdx].Timestamp.Equal(target) {
			nextValue = points[sourceIdx].Value
			hasNext = true
			result[i] = DataPoint{Timestamp: target, Value: nextValue}
			sourceIdx--
			continue
		}

		if hasNext {
			result[i] = DataPoint{Timestamp: target, Value: nextValue}
		}
	}
	return TimeSeries{datapoints: result, label: label}
}

func constantFillSeries(points []DataPoint, targets []time.Time, defaultValue float64, label string) TimeSeries {
	result := make([]DataPoint, len(targets))
	sourceIdx := 0

	for i, target := range targets {
		for sourceIdx < len(points) && points[sourceIdx].Timestamp.Before(target) {
			sourceIdx++
		}

		value := defaultValue
		if sourceIdx < len(points) && points[sourceIdx].Timestamp.Equal(target) {
			value = points[sourceIdx].Value
			sourceIdx++
		}
		result[i] = DataPoint{Timestamp: target, Value: value}
	}

	return TimeSeries{datapoints: result, label: label}
}

func regularGrid(start time.Time, end time.Time, delta time.Duration) []time.Time {
	if delta <= 0 || end.Before(start) {
		return nil
	}

	count := int(end.Sub(start)/delta) + 1
	grid := make([]time.Time, count)
	for i := range grid {
		grid[i] = start.Add(time.Duration(i) * delta)
	}
	return grid
}

func deduplicatedSortedPoints(points []DataPoint) []DataPoint {
	if len(points) == 0 {
		return nil
	}

	working := make([]DataPoint, len(points))
	copy(working, points)
	if !isNonDecreasing(working) {
		slices.SortStableFunc(working, func(left DataPoint, right DataPoint) int {
			return left.Timestamp.Compare(right.Timestamp)
		})
	}

	result := make([]DataPoint, 0, len(working))
	for _, point := range working {
		if len(result) > 0 && result[len(result)-1].Timestamp.Equal(point.Timestamp) {
			result[len(result)-1] = point
			continue
		}
		result = append(result, point)
	}
	return result
}

func nearestPointWithinTolerance(points []DataPoint, timestamps []time.Time, target time.Time, tolerance time.Duration) (DataPoint, bool) {
	idx, found := slices.BinarySearchFunc(timestamps, target, func(left time.Time, right time.Time) int {
		return left.Compare(right)
	})
	if found {
		return points[idx], true
	}

	bestIdx := -1
	bestDistance := tolerance + time.Nanosecond
	if idx < len(points) {
		distance := absDuration(points[idx].Timestamp.Sub(target))
		if distance <= tolerance && distance < bestDistance {
			bestDistance = distance
			bestIdx = idx
		}
	}
	if idx > 0 {
		distance := absDuration(points[idx-1].Timestamp.Sub(target))
		if distance <= tolerance && distance <= bestDistance {
			bestDistance = distance
			bestIdx = idx - 1
		}
	}

	if bestIdx == -1 {
		return DataPoint{}, false
	}
	return points[bestIdx], true
}

func isNonDecreasing(points []DataPoint) bool {
	for i := 1; i < len(points); i++ {
		if points[i].Timestamp.Before(points[i-1].Timestamp) {
			return false
		}
	}
	return true
}

func absDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}
