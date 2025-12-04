package timeseriesgo

import (
	"fmt"
	"time"
)

type DoubleDataPoint struct {
	timestamp  time.Time
	leftValue  float64
	rightValue float64
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
		fmt.Printf("%s, %.2f, %.2f\n", dp.timestamp.Format(time.RFC3339), dp.leftValue, dp.rightValue)
	}
}

func (ts *AlignedSeries) Length() int {
	return len(ts.datapoints)
}

func (ts *AlignedSeries) MapValuesWithReduce(f func(float64, float64) float64) TimeSeries {
	mapped := Empty()
	for _, dp := range ts.datapoints {
		mapped.AddPoint(DataPoint{
			timestamp: dp.timestamp,
			value:     f(dp.leftValue, dp.rightValue),
		})
	}
	return mapped
}
