package generator

import (
	"math"
	"math/rand/v2"
	"testing"
	"time"
)

func TestRenderEvents(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := MakeSeriesIndex(base, time.Hour, 5)

	events := []DemandEvent{
		{StartTime: base.Add(time.Hour), Duration: 2 * time.Hour, Intensity: 2},
		{StartTime: base.Add(2 * time.Hour), Duration: 90 * time.Minute, Intensity: 1},
		{StartTime: base.Add(4 * time.Hour), Duration: 10 * time.Minute, Intensity: 3},
	}

	rendered := RenderEvents(index, events)
	expected := []float64{0, 2, 3, 0.5, 0.5}

	for i, dp := range rendered.DataPoints() {
		if !dp.Timestamp.Equal(index[i]) {
			t.Fatalf("timestamp at index %d: expected %v, got %v", i, index[i], dp.Timestamp)
		}
		if dp.Value != expected[i] {
			t.Fatalf("value at index %d: expected %f, got %f", i, expected[i], dp.Value)
		}
	}
}

func TestRenderEventsScalesShortEventByOverlap(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := MakeSeriesIndex(base, time.Hour, 2)

	rendered := RenderEvents(index, []DemandEvent{
		{StartTime: base.Add(15 * time.Minute), Duration: 30 * time.Minute, Intensity: 2},
	})

	points := rendered.DataPoints()
	if math.Abs(points[0].Value-1.0) > 1e-9 {
		t.Fatalf("expected first bucket average 1.0, got %f", points[0].Value)
	}
	if points[1].Value != 0 {
		t.Fatalf("expected second bucket to stay empty, got %f", points[1].Value)
	}
}

func TestRenderEventsIrregularIndex(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := []time.Time{
		base,
		base.Add(30 * time.Minute),
		base.Add(2 * time.Hour),
	}

	rendered := RenderEvents(index, []DemandEvent{
		{StartTime: base.Add(20 * time.Minute), Duration: time.Hour, Intensity: 1},
	})

	points := rendered.DataPoints()
	if math.Abs(points[0].Value-(10.0/30.0)) > 1e-9 {
		t.Fatalf("expected first irregular bucket average 10/30, got %f", points[0].Value)
	}
	if math.Abs(points[1].Value-(50.0/90.0)) > 1e-9 {
		t.Fatalf("expected second irregular bucket average 50/90, got %f", points[1].Value)
	}
}

func TestRenderEventsEmptyIndex(t *testing.T) {
	if result := RenderEvents(nil, []DemandEvent{{StartTime: time.Now(), Duration: time.Minute, Intensity: 1}}); !result.IsEmpty() {
		t.Fatalf("expected empty series, got length %d", result.Length())
	}
}

func TestEndUseEvents(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	rng := rand.New(rand.NewPCG(3, 5))

	events := EndUseEvents(start, end, EndUseConfig{
		EventsPerDayMean: 3,
		DurationMean:     10 * time.Minute,
		IntensityMean:    0.2,
		Rand:             rng,
	})

	if len(events) == 0 {
		t.Fatal("expected generated events, got none")
	}

	for i, event := range events {
		if event.StartTime.Before(start) || !event.StartTime.Before(end) {
			t.Fatalf("event %d start time out of range: %+v", i, event)
		}
		if event.Duration <= 0 {
			t.Fatalf("event %d has non-positive duration: %+v", i, event)
		}
		if event.Intensity <= 0 {
			t.Fatalf("event %d has non-positive intensity: %+v", i, event)
		}
		if event.StartTime.Add(event.Duration).After(end) {
			t.Fatalf("event %d extends past the requested end: %+v", i, event)
		}
		if i > 0 && events[i-1].StartTime.After(event.StartTime) {
			t.Fatalf("events are not sorted: %v", events)
		}
	}
}

func TestEndUseEventsWithSameSeedAreReproducible(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	cfgA := EndUseConfig{
		EventsPerDayMean: 4,
		DurationMean:     12 * time.Minute,
		IntensityMean:    0.15,
		Rand:             rand.New(rand.NewPCG(11, 17)),
	}
	cfgB := EndUseConfig{
		EventsPerDayMean: 4,
		DurationMean:     12 * time.Minute,
		IntensityMean:    0.15,
		Rand:             rand.New(rand.NewPCG(11, 17)),
	}

	left := EndUseEvents(start, end, cfgA)
	right := EndUseEvents(start, end, cfgB)
	if len(left) != len(right) {
		t.Fatalf("expected equal event counts, got %d and %d", len(left), len(right))
	}
	for i := range left {
		if left[i] != right[i] {
			t.Fatalf("event %d mismatch: %+v vs %+v", i, left[i], right[i])
		}
	}
}

func TestEndUseEventsInvalidInput(t *testing.T) {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	cases := [][]DemandEvent{
		EndUseEvents(end, start, EndUseConfig{EventsPerDayMean: 1, DurationMean: time.Minute, IntensityMean: 1}),
		EndUseEvents(start, end, EndUseConfig{EventsPerDayMean: 0, DurationMean: time.Minute, IntensityMean: 1}),
		EndUseEvents(start, end, EndUseConfig{EventsPerDayMean: 1, DurationMean: 0, IntensityMean: 1}),
		EndUseEvents(start, end, EndUseConfig{EventsPerDayMean: 1, DurationMean: time.Minute, IntensityMean: 0}),
	}

	for i, events := range cases {
		if len(events) != 0 {
			t.Fatalf("case %d: expected no events, got %v", i, events)
		}
	}
}

func TestHouseholdDemand(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := MakeSeriesIndex(base, 15*time.Minute, 24*7*4)

	oneOccupant := HouseholdDemand(index, HouseholdDemandConfig{Occupants: 1, Rand: rand.New(rand.NewPCG(1, 2))})
	fourOccupants := HouseholdDemand(index, HouseholdDemandConfig{Occupants: 4, Rand: rand.New(rand.NewPCG(1, 2))})

	if oneOccupant.Length() != len(index) || fourOccupants.Length() != len(index) {
		t.Fatalf("expected preserved index length, got %d and %d", oneOccupant.Length(), fourOccupants.Length())
	}

	totalOne := 0.0
	for i, dp := range oneOccupant.DataPoints() {
		if !dp.Timestamp.Equal(index[i]) {
			t.Fatalf("one-occupant timestamp at index %d: expected %v, got %v", i, index[i], dp.Timestamp)
		}
		if dp.Value < 0 {
			t.Fatalf("one-occupant series contains negative value at index %d: %f", i, dp.Value)
		}
		totalOne += dp.Value
	}

	totalFour := 0.0
	for i, dp := range fourOccupants.DataPoints() {
		if !dp.Timestamp.Equal(index[i]) {
			t.Fatalf("four-occupant timestamp at index %d: expected %v, got %v", i, index[i], dp.Timestamp)
		}
		if dp.Value < 0 {
			t.Fatalf("four-occupant series contains negative value at index %d: %f", i, dp.Value)
		}
		totalFour += dp.Value
	}

	if totalFour <= totalOne {
		t.Fatalf("expected four occupants to generate more total demand, got one=%f four=%f", totalOne, totalFour)
	}
}

func TestHouseholdDemandShowsDiurnalPattern(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := MakeSeriesIndex(base, 15*time.Minute, 24*4*30)

	household := HouseholdDemand(index, HouseholdDemandConfig{Occupants: 3, Rand: rand.New(rand.NewPCG(21, 34))})

	nightTotal := 0.0
	nightCount := 0
	peakTotal := 0.0
	peakCount := 0

	for _, dp := range household.DataPoints() {
		hour := dp.Timestamp.Hour()
		switch {
		case hour >= 1 && hour < 5:
			nightTotal += dp.Value
			nightCount++
		case hour >= 6 && hour < 9:
			peakTotal += dp.Value
			peakCount++
		case hour >= 19 && hour < 22:
			peakTotal += dp.Value
			peakCount++
		}
	}

	nightMean := nightTotal / float64(nightCount)
	peakMean := peakTotal / float64(peakCount)
	if peakMean <= nightMean {
		t.Fatalf("expected daytime peaks to exceed night demand, got peak=%f night=%f", peakMean, nightMean)
	}
}

func TestHouseholdDemandWithSameSeedIsReproducible(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := MakeSeriesIndex(base, 15*time.Minute, 24*4*7)

	left := HouseholdDemand(index, HouseholdDemandConfig{Occupants: 3, Rand: rand.New(rand.NewPCG(13, 21))})
	right := HouseholdDemand(index, HouseholdDemandConfig{Occupants: 3, Rand: rand.New(rand.NewPCG(13, 21))})

	leftPoints := left.DataPoints()
	rightPoints := right.DataPoints()
	for i := range leftPoints {
		if leftPoints[i] != rightPoints[i] {
			t.Fatalf("expected reproducible household demand, mismatch at %d: %+v vs %+v", i, leftPoints[i], rightPoints[i])
		}
	}
}

func TestHouseholdDemandInvalidInput(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := MakeSeriesIndex(base, time.Hour, 4)

	if result := HouseholdDemand(nil, HouseholdDemandConfig{Occupants: 2}); !result.IsEmpty() {
		t.Fatalf("expected empty series for empty index, got length %d", result.Length())
	}
	if result := HouseholdDemand(index, HouseholdDemandConfig{Occupants: 0}); !result.IsEmpty() {
		t.Fatalf("expected empty series for invalid occupants, got length %d", result.Length())
	}
}
