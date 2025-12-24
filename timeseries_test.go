package timeseriesgo

import (
	"math"
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
	return dp.Value > 15.0
}

func sum(dps []DataPoint) float64 {
	total := 0.0
	for _, dp := range dps {
		total += dp.Value
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

	if minVal.Value != 5.0 {
		t.Errorf("Expected min value 5.0, got %f", minVal.Value)
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

	if maxVal.Value != 20.0 {
		t.Errorf("Expected max value 20.0, got %f", maxVal.Value)
	}
}

func TestSlice(t *testing.T) {
	ts := Empty()
	now := time.Now()

	start := now.Add(2 * time.Minute)
	end := now.Add(6 * time.Minute)

	ts.AddPoint(DataPoint{Timestamp: now, Value: 1.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(time.Minute), Value: -3.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(2 * time.Minute), Value: 6.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(3 * time.Minute), Value: 6.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(4 * time.Minute), Value: 6.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(5 * time.Minute), Value: 8.0})
	res := ts.Slice(start, end)
	if res.Length() != 4 {
		t.Errorf("Expected sliced TimeSeries length 4, got %d", res.Length())
	}

	if res.DataPoints()[0].Timestamp != start {
		t.Errorf("Expected first datapoint timestamp %v, got %v", start, res.DataPoints()[0].Timestamp)
	}

	if res.DataPoints()[3].Timestamp != now.Add(5*time.Minute) {
		t.Errorf("Expected last datapoint timestamp %v, got %v", now.Add(5*time.Minute), res.DataPoints()[3].Timestamp)
	}
}

func TestResolution(t *testing.T) {
	ts := Empty()
	base := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)

	ts.AddPoint(DataPoint{Timestamp: base, Value: 1})
	ts.AddPoint(DataPoint{Timestamp: base.Add(time.Minute), Value: 2})
	ts.AddPoint(DataPoint{Timestamp: base.Add(2 * time.Minute), Value: 3})
	ts.AddPoint(DataPoint{Timestamp: base.Add(4 * time.Minute), Value: 4})

	res, err := ts.Resolution()
	if err != nil {
		t.Fatalf("Unexpected error computing resolution: %v", err)
	}

	if res != time.Minute {
		t.Errorf("Expected resolution 1m, got %v", res)
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
		return dp.Timestamp.Equal(now.Add(5 * time.Minute))
	})
	expectedValues := []float64{5.0}
	if mapped.Length() != len(expectedValues) {
		t.Errorf("Expected filtered TimeSeries length %d, got %d", len(expectedValues), mapped.Length())
	}
}

func TestJoin(t *testing.T) {
	ts1 := Empty()
	ts2 := Empty()

	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 10.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 20.0})

	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 15.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 25.0})

	expectedPoints := []DoubleDataPoint{
		{Timestamp: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), LeftValue: 10.0, RightValue: 15.0},
	}

	joined := ts1.Join(ts2)

	if joined.Length() != len(expectedPoints) {
		t.Errorf("Expected joined AlignedSeries length %d, got %d", len(expectedPoints), joined.Length())
	}
	for i, dp := range joined.DataPoints() {
		expDp := expectedPoints[i]
		if !dp.Timestamp.Equal(expDp.Timestamp) || dp.LeftValue != expDp.LeftValue || dp.RightValue != expDp.RightValue {
			t.Errorf("At index %d, expected datapoint %+v, got %+v", i, expDp, dp)
		}
	}
}

func TestJoinLeft(t *testing.T) {
	ts1 := Empty()
	ts2 := Empty()

	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 10.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 20.0})

	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 15.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 25.0})

	expectedPoints := []DoubleDataPoint{
		{Timestamp: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), LeftValue: 10.0, RightValue: 15.0},
		{Timestamp: time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), LeftValue: 20.0, RightValue: 0.0},
	}

	joined := ts1.JoinLeft(ts2, 0.0)

	if joined.Length() != len(expectedPoints) {
		t.Errorf("Expected joined AlignedSeries length %d, got %d", len(expectedPoints), joined.Length())
	}
	for i, dp := range joined.DataPoints() {
		expDp := expectedPoints[i]
		if !dp.Timestamp.Equal(expDp.Timestamp) || dp.LeftValue != expDp.LeftValue || dp.RightValue != expDp.RightValue {
			t.Errorf("At index %d, expected datapoint %+v, got %+v", i, expDp, dp)
		}
	}
}

func TestJoinOuter(t *testing.T) {
	ts1 := Empty()
	ts2 := Empty()

	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 10.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 20.0})

	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 15.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 25.0})

	expectedPoints := []DoubleDataPoint{
		{Timestamp: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), LeftValue: 10.0, RightValue: 15.0},
		{Timestamp: time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), LeftValue: 20.0, RightValue: 0.0},
		{Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), LeftValue: 0.0, RightValue: 25.0},
	}

	joined := ts1.JoinOuter(ts2, 0.0, 0.0)

	if joined.Length() != len(expectedPoints) {
		t.Errorf("Expected joined AlignedSeries length %d, got %d", len(expectedPoints), joined.Length())
	}
	for i, dp := range joined.DataPoints() {
		expDp := expectedPoints[i]
		if !dp.Timestamp.Equal(expDp.Timestamp) || dp.LeftValue != expDp.LeftValue || dp.RightValue != expDp.RightValue {
			t.Errorf("At index %d, expected datapoint %+v, got %+v", i, expDp, dp)
		}
	}
}

func TestDifferentiate(t *testing.T) {
	ts := Empty()
	now := time.Now()

	ts.AddPoint(DataPoint{Timestamp: now, Value: 2.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(time.Minute), Value: -4.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(2 * time.Minute), Value: -6.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(3 * time.Minute), Value: 8.0})

	res := ts.Differentiate()

	expected := []DataPoint{
		{Timestamp: now.Add(time.Minute), Value: -6},
		{Timestamp: now.Add(2 * time.Minute), Value: -2},
		{Timestamp: now.Add(3 * time.Minute), Value: 14},
	}
	if res.Length() != len(expected) {
		t.Fatalf("expected %d points, got %d", len(expected), res.Length())
	}
	for i, dp := range res.DataPoints() {
		if !dp.Timestamp.Equal(expected[i].Timestamp) || dp.Value != expected[i].Value {
			t.Errorf("idx %d expected %+v, got %+v", i, expected[i], dp)
		}
	}
}

func TestIntegrate(t *testing.T) {
	ts := Empty()
	now := time.Now()

	ts.AddPoint(DataPoint{Timestamp: now, Value: 2.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(time.Minute), Value: -4.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(2 * time.Minute), Value: -6.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(3 * time.Minute), Value: 8.0})

	res := ts.Integrate()

	expected := []DataPoint{
		{Timestamp: now.Add(time.Minute), Value: -2},
		{Timestamp: now.Add(2 * time.Minute), Value: -10},
		{Timestamp: now.Add(3 * time.Minute), Value: 2},
	}
	if res.Length() != len(expected) {
		t.Fatalf("expected %d points, got %d", len(expected), res.Length())
	}
	for i, dp := range res.DataPoints() {
		if !dp.Timestamp.Equal(expected[i].Timestamp) || dp.Value != expected[i].Value {
			t.Errorf("idx %d expected %+v, got %+v", i, expected[i], dp)
		}
	}
}

func TestMedian(t *testing.T) {
	ts := Empty()
	now := time.Now()

	ts.AddPoint(DataPoint{Timestamp: now, Value: 1.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(time.Minute), Value: 2})
	ts.AddPoint(DataPoint{Timestamp: now.Add(2 * time.Minute), Value: 3.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(3 * time.Minute), Value: 4.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(4 * time.Minute), Value: 5.0})

	m, err := ts.Median()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if m != 3.0 {
		t.Errorf("Expected median value was %f, got ", m)
	}

	ts.AddPoint(DataPoint{Timestamp: now.Add(5 * time.Minute), Value: 6.0})

	m2, err2 := ts.Median()

	if err2 != nil {
		t.Errorf("Unexpected error: %v", err2)
	}

	if m2 != 3.5 {
		t.Errorf("Expected median value was %f, got ", m2)
	}
}

func TestRollingWindow(t *testing.T) {
	ts := Empty()
	expected := Empty()
	now := time.Now()

	ts.AddPoint(DataPoint{Timestamp: now, Value: 1.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(10 * time.Minute), Value: 2})
	ts.AddPoint(DataPoint{Timestamp: now.Add(30 * time.Minute), Value: 3.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(50 * time.Minute), Value: 4.0})
	ts.AddPoint(DataPoint{Timestamp: now.Add(80 * time.Minute), Value: 5.0})

	expected.AddPoint(DataPoint{Timestamp: now, Value: 1.0})
	expected.AddPoint(DataPoint{Timestamp: now.Add(10 * time.Minute), Value: 3.0})
	expected.AddPoint(DataPoint{Timestamp: now.Add(30 * time.Minute), Value: 6.0})
	expected.AddPoint(DataPoint{Timestamp: now.Add(50 * time.Minute), Value: 10.0})
	expected.AddPoint(DataPoint{Timestamp: now.Add(80 * time.Minute), Value: 12.0})

	rwts := ts.RollingWindow(time.Hour, func(vs []float64) float64 {
		res := 0.0
		for _, v := range vs {
			res += v
		}
		return res
	})

	if rwts.Length() != len(expected.DataPoints()) {
		t.Errorf("Expected length %d, got %d", len(expected.DataPoints()), rwts.Length())
	}
	for i, dp := range rwts.DataPoints() {
		expDp := expected.DataPoints()[i]
		if !dp.Timestamp.Equal(expDp.Timestamp) || dp.Value != expDp.Value {
			t.Errorf("At index %d, expected datapoint %+v, got %+v", i, expDp, dp)
		}
	}
}

func TestResampleWithDefaultValueFillsMissing(t *testing.T) {
	base := time.Unix(0, 0)
	ts := Empty()
	ts.AddPoint(DataPoint{Timestamp: base.Add(1 * time.Second), Value: 1})
	ts.AddPoint(DataPoint{Timestamp: base.Add(2 * time.Second), Value: 2})
	ts.AddPoint(DataPoint{Timestamp: base.Add(3 * time.Second), Value: 3})
	ts.AddPoint(DataPoint{Timestamp: base.Add(5 * time.Second), Value: 5})

	resampled := ts.ResampleWithDefaultValue(1*time.Second, 1)

	expectedTimes := []time.Time{
		base.Add(1 * time.Second),
		base.Add(2 * time.Second),
		base.Add(3 * time.Second),
		base.Add(4 * time.Second),
		base.Add(5 * time.Second),
	}
	expectedValues := []float64{1, 2, 3, 1, 5}

	if resampled.Length() != len(expectedTimes) {
		t.Fatalf("expected %d points after resample, got %d", len(expectedTimes), resampled.Length())
	}

	for i, dp := range resampled.DataPoints() {
		if !dp.Timestamp.Equal(expectedTimes[i]) {
			t.Errorf("idx %d expected timestamp %v, got %v", i, expectedTimes[i], dp.Timestamp)
		}
		if dp.Value != expectedValues[i] {
			t.Errorf("idx %d expected value %v, got %v", i, expectedValues[i], dp.Value)
		}
	}
}

func TestResampleZeroDeltaReturnsOriginal(t *testing.T) {
	base := time.Unix(0, 0)
	ts := Empty()
	ts.AddPoint(DataPoint{Timestamp: base, Value: 1})
	ts.AddPoint(DataPoint{Timestamp: base.Add(time.Second), Value: 3})

	res := ts.Resample(0, func(d1 DataPoint, d2 DataPoint, idx time.Time) float64 {
		return 2
	})

	if res.Length() != ts.Length() {
		t.Fatalf("expected resampled length %d, got %d", ts.Length(), res.Length())
	}
	for i, dp := range res.DataPoints() {
		orig := ts.DataPoints()[i]
		if !dp.Timestamp.Equal(orig.Timestamp) || dp.Value != orig.Value {
			t.Errorf("idx %d expected %+v, got %+v", i, orig, dp)
		}
	}
}

func TestResampleEmptySeriesReturnsEmpty(t *testing.T) {
	ts := Empty()
	res := ts.Resample(1*time.Second, func(d1 DataPoint, d2 DataPoint, idx time.Time) float64 {
		return 0
	})
	if !res.IsEmpty() {
		t.Fatalf("expected empty result for empty input, got length %d", res.Length())
	}
}

func TestResampleLargeDeltaKeepsOriginalPoints(t *testing.T) {
	base := time.Unix(0, 0)
	ts := Empty()
	ts.AddPoint(DataPoint{Timestamp: base.Add(1 * time.Second), Value: 1})
	ts.AddPoint(DataPoint{Timestamp: base.Add(3 * time.Second), Value: 3})
	ts.AddPoint(DataPoint{Timestamp: base.Add(4 * time.Second), Value: 4})

	res := ts.Resample(10*time.Second, func(d1 DataPoint, d2 DataPoint, idx time.Time) float64 {
		return 999 // should not be used
	})

	if res.Length() != 1 {
		t.Fatalf("expected only the first grid-aligned point, got %d", res.Length())
	}
	if !res.DataPoints()[0].Timestamp.Equal(base.Add(1*time.Second)) || res.DataPoints()[0].Value != 1 {
		t.Fatalf("unexpected point %+v", res.DataPoints()[0])
	}
}

func TestResampleWithInterpolationFunction(t *testing.T) {
	base := time.Unix(0, 0)
	ts := Empty()
	ts.AddPoint(DataPoint{Timestamp: base.Add(0 * time.Second), Value: 0})
	ts.AddPoint(DataPoint{Timestamp: base.Add(2 * time.Second), Value: 2})

	res := ts.Resample(1*time.Second, func(d1 DataPoint, d2 DataPoint, idx time.Time) float64 {
		// linear interpolation between d1 (previous) and d2 (next)
		total := d2.Timestamp.Sub(d1.Timestamp).Seconds()
		elapsed := idx.Sub(d1.Timestamp).Seconds()
		return d1.Value + (d2.Value-d1.Value)*(elapsed/total)
	})

	expected := []DataPoint{
		{Timestamp: base.Add(0 * time.Second), Value: 0},
		{Timestamp: base.Add(1 * time.Second), Value: 1},
		{Timestamp: base.Add(2 * time.Second), Value: 2},
	}

	if res.Length() != len(expected) {
		t.Fatalf("expected %d points, got %d", len(expected), res.Length())
	}
	for i, dp := range res.DataPoints() {
		if !dp.Timestamp.Equal(expected[i].Timestamp) || math.Abs(dp.Value-expected[i].Value) > 1e-9 {
			t.Errorf("idx %d expected %+v, got %+v", i, expected[i], dp)
		}
	}
}

func TestResampleSinglePointDeltaPositiveReturnsCopy(t *testing.T) {
	base := time.Unix(0, 0)
	ts := Empty()
	ts.AddPoint(DataPoint{Timestamp: base, Value: 5})

	res := ts.Resample(2*time.Second, func(d1 DataPoint, d2 DataPoint, idx time.Time) float64 {
		return 0
	})

	if res.Length() != ts.Length() {
		t.Fatalf("expected length %d, got %d", ts.Length(), res.Length())
	}
	if !res.DataPoints()[0].Timestamp.Equal(base) || res.DataPoints()[0].Value != 5 {
		t.Fatalf("expected original point preserved, got %+v", res.DataPoints()[0])
	}
}

func TestResampleGapLargerThanDeltaInsertsAllGridPoints(t *testing.T) {
	base := time.Unix(0, 0)
	ts := Empty()
	ts.AddPoint(DataPoint{Timestamp: base, Value: 0})
	ts.AddPoint(DataPoint{Timestamp: base.Add(10 * time.Second), Value: 10})

	res := ts.Resample(2*time.Second, func(d1 DataPoint, d2 DataPoint, idx time.Time) float64 {
		// linear interpolation
		total := d2.Timestamp.Sub(d1.Timestamp).Seconds()
		elapsed := idx.Sub(d1.Timestamp).Seconds()
		return d1.Value + (d2.Value-d1.Value)*(elapsed/total)
	})

	expectedValues := []float64{0, 2, 4, 6, 8, 10}
	if res.Length() != len(expectedValues) {
		t.Fatalf("expected %d points, got %d", len(expectedValues), res.Length())
	}
	for i, dp := range res.DataPoints() {
		if dp.Value != expectedValues[i] {
			t.Errorf("idx %d expected value %.0f, got %.0f", i, expectedValues[i], dp.Value)
		}
	}
}

func TestInterpolateLinear(t *testing.T) {
	base := time.Unix(0, 0)
	ts := Empty()
	ts.AddPoint(DataPoint{Timestamp: base, Value: 0})
	ts.AddPoint(DataPoint{Timestamp: base.Add(2 * time.Second), Value: 2})

	res := ts.Interpolate(1 * time.Second)

	expected := []DataPoint{
		{Timestamp: base, Value: 0},
		{Timestamp: base.Add(1 * time.Second), Value: 1},
		{Timestamp: base.Add(2 * time.Second), Value: 2},
	}
	if res.Length() != len(expected) {
		t.Fatalf("expected %d points, got %d", len(expected), res.Length())
	}
	for i, dp := range res.DataPoints() {
		if !dp.Timestamp.Equal(expected[i].Timestamp) || math.Abs(dp.Value-expected[i].Value) > 1e-9 {
			t.Errorf("idx %d expected %+v, got %+v", i, expected[i], dp)
		}
	}
}

func TestStepCarriesForward(t *testing.T) {
	base := time.Unix(0, 0)
	ts := Empty()
	ts.AddPoint(DataPoint{Timestamp: base, Value: 5})
	ts.AddPoint(DataPoint{Timestamp: base.Add(4 * time.Second), Value: 7})

	res := ts.Step(2 * time.Second)

	expected := []DataPoint{
		{Timestamp: base.Add(2 * time.Second), Value: 2.5}, // 5 split into 2 steps
		{Timestamp: base.Add(4 * time.Second), Value: 2.5},
	}
	if res.Length() != len(expected) {
		t.Fatalf("expected %d points, got %d", len(expected), res.Length())
	}
	for i, dp := range res.DataPoints() {
		if !dp.Timestamp.Equal(expected[i].Timestamp) || dp.Value != expected[i].Value {
			t.Errorf("idx %d expected %+v, got %+v", i, expected[i], dp)
		}
	}
}
