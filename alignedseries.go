package timeseriesgo

import (
	"fmt"
	"time"
)

type DoubleDataPoint struct {
	Timestamp  time.Time
	LeftValue  float64
	RightValue float64
}

type AlignedSeries struct {
	datapoints []DoubleDataPoint
	label      string
}

/**
 * Creates an empty AlignedSeries with the provided label.
 *
 * @param label The label to assign to the created aligned series.
 *
 * @return A new empty AlignedSeries.
 */
func EmptyLabeledAlignedSeries(label string) AlignedSeries {
	return AlignedSeries{datapoints: []DoubleDataPoint{}, label: label}
}

/**
 * Prints the AlignedSeries in a human-readable format.
 *
 * @return None. The function writes the aligned series to standard output.
 */
func (ts *AlignedSeries) Print() {
	fmt.Println("Timestamp, Left Value, Right Value")
	for _, dp := range ts.datapoints {
		fmt.Printf("%s, %.2f, %.2f\n", dp.Timestamp.Format(time.RFC3339), dp.LeftValue, dp.RightValue)
	}
}

/**
 * Returns the number of aligned datapoints in the series.
 *
 * @return The number of DoubleDataPoint entries stored in the series.
 */
func (ts *AlignedSeries) Length() int {
	return len(ts.datapoints)
}

/**
 * Maps a reducer function over aligned left and right values.
 *
 * @param f A function that takes the left and right values and returns a single float64 result.
 *
 * @return A new TimeSeries with the reduced values at the original aligned timestamps.
 */
func (ts *AlignedSeries) MapValuesWithReduce(f func(float64, float64) float64) TimeSeries {
	mapped := Empty()
	for _, dp := range ts.datapoints {
		mapped.AddPoint(DataPoint{
			Timestamp: dp.Timestamp,
			Value:     f(dp.LeftValue, dp.RightValue),
		})
	}
	return mapped
}

/**
 * Returns the aligned datapoints of the series.
 *
 * @return A shallow copy of the underlying DoubleDataPoint slice for safe read-only access.
 */
func (ts *AlignedSeries) DataPoints() []DoubleDataPoint {
	cp := make([]DoubleDataPoint, len(ts.datapoints))
	copy(cp, ts.datapoints)
	return cp
}
