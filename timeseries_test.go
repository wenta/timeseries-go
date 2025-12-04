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
