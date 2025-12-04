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

func (ts *TimeSeries) AddPoint(dp DataPoint) {
	ts.datapoints = append(ts.datapoints, dp)
}

func (ts *TimeSeries) Print() {
	fmt.Println("Timestamp, Value")
	for _, dp := range ts.datapoints {
		fmt.Printf("%s, %.2f\n", dp.timestamp.Format(time.RFC3339), dp.value)
	}
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
