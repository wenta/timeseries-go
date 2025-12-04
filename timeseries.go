package timeseriesgo

import (
	"errors"
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
