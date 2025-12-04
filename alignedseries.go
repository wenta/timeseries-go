package timeseriesgo

import "time"

type DoubleDataPoint struct {
	timestamp  time.Time
	leftValue  float64
	rightValue float64
}

type AlignedSeries struct {
	datapoints []DoubleDataPoint
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
