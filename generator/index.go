package generator

import "time"

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
	index := make([]time.Time, 0, count)
	for i := 0; i < count; i++ {
		index = append(index, start.Add(time.Duration(i)*interval))
	}
	return index
}
