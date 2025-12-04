package timeseriesgo

import (
	"testing"
	"time"
)

func TestCreateTimeSeries(t *testing.T) {
	ts := Empty()
	if ts.Length() != 0 {
		t.Errorf("Empty TimeSeries should have length 0, got %d", ts.Length())
	}
}

func TestAddPoint(t *testing.T) {
	ts := Empty()
	expected := []float64{10.20}
	ts.AddPoint(DataPoint{time.Now(), 10.20})
	if ts.Length() != 1 {
		t.Errorf("Expected one datapoint")
	}

	if ts.Values()[0] != expected[0] {
		t.Errorf("Expected one datapoint")
	}
}

func roundToHour(dt time.Time) time.Time {
	return time.Date(dt.Year(), dt.Month(), dt.Day(), dt.Hour(), 0, 0, 0, dt.Location())
}

func increment(x float64) float64 {
	return x + 1.0
}

func greaterThan15(dp DataPoint) bool {
	return dp.value > 15.0
}

func sum(dps []DataPoint) float64 {
	total := 0.0
	for _, dp := range dps {
		total += dp.value
	}
	return total
}

func TestGroupByTime(t *testing.T) {
	ts := Empty()
	expected := Empty()
	ts.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 10.0})
	ts.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 30, 0, 0, time.UTC), 20.0})
	ts.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 30.0})
	ts.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 30, 0, 0, time.UTC), 10.0})
	ts.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 20.0})
	ts.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 30, 0, 0, time.UTC), 30.0})
	expected.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 30.0})
	expected.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 40.0})
	expected.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 50.0})

	grouped := ts.GroupByTime(roundToHour, sum)
	if grouped.IsEmpty() {
		t.Errorf("Expected non-empty grouped TimeSeries")
	}
	if grouped.Length() != expected.Length() {
		t.Errorf("Expected grouped TimeSeries length %d, got %d", expected.Length(), grouped.Length())
	}
	for i, val := range grouped.Values() {
		if val != expected.Values()[i] {
			t.Errorf("At index %d, expected value %f, got %f", i, expected.Values()[i], val)
		}
	}
}

func TestMerge(t *testing.T) {
	ts1 := Empty()
	ts2 := Empty()
	expected := Empty()

	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 10.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 20.0})

	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 30, 0, 0, time.UTC), 15.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 25.0})

	expected.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 10.0})
	expected.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 30, 0, 0, time.UTC), 15.0})
	expected.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 20.0})
	expected.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 25.0})

	merged := ts1.Merge(ts2)

	if merged.IsEmpty() {
		t.Errorf("Expected non-empty merged TimeSeries")
	}
	if merged.Length() != expected.Length() {
		t.Errorf("Expected merged TimeSeries length %d, got %d", expected.Length(), merged.Length())
	}
	for i, val := range merged.Values() {
		if val != expected.Values()[i] {
			t.Errorf("At index %d, expected value %f, got %f", i, expected.Values()[i], val)
		}
	}
}

func TestMerge_BiggerSeriesWithSmaller(t *testing.T) {
	ts1 := Empty()
	ts2 := Empty()
	expected := Empty()

	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 10.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 20.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 30.0})

	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 30, 0, 0, time.UTC), 15.0})

	expected.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 10.0})
	expected.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 20.0})
	expected.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 30, 0, 0, time.UTC), 15.0})
	expected.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 30.0})

	merged := ts1.Merge(ts2)

	if merged.IsEmpty() {
		t.Errorf("Expected non-empty merged TimeSeries")
	}
	if merged.Length() != expected.Length() {
		t.Errorf("Expected merged TimeSeries length %d, got %d", expected.Length(), merged.Length())
	}
	for i, val := range merged.Values() {
		if val != expected.Values()[i] {
			t.Errorf("At index %d, expected value %f, got %f", i, expected.Values()[i], val)
		}
	}
}

/**
 * Tests for statistics functions
 */

func TestMin(t *testing.T) {
	ts := Empty()
	now := time.Now()
	ts.AddPoint(DataPoint{now, 10.0})
	ts.AddPoint(DataPoint{now.Add(5 * time.Minute), 5.0})
	ts.AddPoint(DataPoint{now.Add(10 * time.Minute), 20.0})

	minVal, err := ts.Min()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if minVal.value != 5.0 {
		t.Errorf("Expected min value 5.0, got %f", minVal.value)
	}
}

func TestMax(t *testing.T) {
	ts := Empty()
	now := time.Now()
	ts.AddPoint(DataPoint{now, 10.0})
	ts.AddPoint(DataPoint{now.Add(5 * time.Minute), 5.0})
	ts.AddPoint(DataPoint{now.Add(10 * time.Minute), 20.0})

	maxVal, err := ts.Max()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if maxVal.value != 20.0 {
		t.Errorf("Expected max value 20.0, got %f", maxVal.value)
	}
}

func TestSlice(t *testing.T) {
	ts := Empty()
	now := time.Now()

	start := now.Add(2 * time.Minute)
	end := now.Add(6 * time.Minute)

	ts.AddPoint(DataPoint{timestamp: now, value: 1.0})
	ts.AddPoint(DataPoint{timestamp: now.Add(time.Minute), value: -3.0})
	ts.AddPoint(DataPoint{timestamp: now.Add(2 * time.Minute), value: 6.0})
	ts.AddPoint(DataPoint{timestamp: now.Add(3 * time.Minute), value: 6.0})
	ts.AddPoint(DataPoint{timestamp: now.Add(4 * time.Minute), value: 6.0})
	ts.AddPoint(DataPoint{timestamp: now.Add(5 * time.Minute), value: 8.0})
	res := ts.Slice(start, end)
	if res.Length() != 4 {
		t.Errorf("Expected sliced TimeSeries length 4, got %d", res.Length())
	}

	if res.datapoints[0].timestamp != start {
		t.Errorf("Expected first datapoint timestamp %v, got %v", start, res.datapoints[0].timestamp)
	}

	if res.datapoints[3].timestamp != now.Add(5*time.Minute) {
		t.Errorf("Expected last datapoint timestamp %v, got %v", now.Add(5*time.Minute), res.datapoints[3].timestamp)
	}
}

func TestMap(t *testing.T) {
	ts := Empty()
	now := time.Now()
	ts.AddPoint(DataPoint{now, 10.0})
	ts.AddPoint(DataPoint{now.Add(5 * time.Minute), 5.0})
	ts.AddPoint(DataPoint{now.Add(10 * time.Minute), 20.0})

	mapped := ts.MapValues(increment)

	expectedValues := []float64{11.0, 6.0, 21.0}
	for i, val := range mapped.Values() {
		if val != expectedValues[i] {
			t.Errorf("At index %d, expected mapped value %f, got %f", i, expectedValues[i], val)
		}
	}
}

func TestFilter(t *testing.T) {
	ts := Empty()
	now := time.Now()
	ts.AddPoint(DataPoint{now, 10.0})
	ts.AddPoint(DataPoint{now.Add(5 * time.Minute), 5.0})
	ts.AddPoint(DataPoint{now.Add(10 * time.Minute), 20.0})

	mapped := ts.Filter(greaterThan15)
	expectedValues := []float64{20.0}
	if mapped.Length() != len(expectedValues) {
		t.Errorf("Expected filtered TimeSeries length %d, got %d", len(expectedValues), mapped.Length())
	}
	for i, val := range mapped.Values() {
		if val != expectedValues[i] {
			t.Errorf("At index %d, expected filtered value %f, got %f", i, expectedValues[i], val)
		}
	}
}

func TestFilter_ByIndex(t *testing.T) {
	ts := Empty()
	now := time.Now()
	ts.AddPoint(DataPoint{now, 10.0})
	ts.AddPoint(DataPoint{now.Add(5 * time.Minute), 5.0})
	ts.AddPoint(DataPoint{now.Add(10 * time.Minute), 20.0})

	mapped := ts.Filter(func(dp DataPoint) bool {
		return dp.timestamp.Equal(now.Add(5 * time.Minute))
	})
	expectedValues := []float64{5.0}
	if mapped.Length() != len(expectedValues) {
		t.Errorf("Expected filtered TimeSeries length %d, got %d", len(expectedValues), mapped.Length())
	}
}
