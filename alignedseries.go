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
}

/**
 * Prints the TimeSeries in a human-readable format.
 */
func (ts *AlignedSeries) Print() {
	fmt.Println("Timestamp, Left Value, Right Value")
	for _, dp := range ts.datapoints {
		fmt.Printf("%s, %.2f, %.2f\n", dp.Timestamp.Format(time.RFC3339), dp.LeftValue, dp.RightValue)
	}
}

func (ts *AlignedSeries) Length() int {
	return len(ts.datapoints)
}

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

// DataPoints returns a shallow copy of the underlying aligned datapoints for read-only use.
func (ts *AlignedSeries) DataPoints() []DoubleDataPoint {
	cp := make([]DoubleDataPoint, len(ts.datapoints))
	copy(cp, ts.datapoints)
	return cp
}
