package generator

import (
	"math"
	"math/rand/v2"
	"slices"
	"sort"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

// DemandEvent describes a demand contribution active over a time interval.
type DemandEvent struct {
	// StartTime is the inclusive start time of the event.
	StartTime time.Time
	// Duration is the active duration of the event.
	Duration time.Duration
	// Intensity is the value contributed while the event is active.
	Intensity float64
}

// EndUseConfig controls stochastic event generation for one end-use category.
type EndUseConfig struct {
	// EventsPerDayMean is the average number of events expected per day.
	EventsPerDayMean float64
	// DurationMean is the mean duration of each event.
	DurationMean time.Duration
	// IntensityMean is the mean intensity contributed by each event.
	IntensityMean float64
	// Rand is the random source used for event generation. Nil uses the package default RNG.
	Rand *rand.Rand
}

// HouseholdDemandConfig controls a simple stochastic household demand generator.
type HouseholdDemandConfig struct {
	// Occupants is the number of household occupants used to scale activity.
	Occupants int
	// Rand is the random source used for synthetic demand generation. Nil uses the package default RNG.
	Rand *rand.Rand
}

/**
 * Renders demand events onto a fixed timestamp index as a zero-baseline TimeSeries.
 *
 * Intensity is treated as an active contribution rate. Each output point stores
 * the average contribution over its bucket, computed from the overlap between the
 * event interval and the bucket interval. Overlapping events add together.
 *
 * @param index The timestamps used for the output series.
 * @param events The demand events to render.
 *
 * @return A TimeSeries with timestamps from index and average event contributions over each bucket.
 */
func RenderEvents(index []time.Time, events []DemandEvent) timeseriesgo.TimeSeries {
	if len(index) == 0 {
		return timeseriesgo.Empty()
	}

	values := make([]float64, len(index))
	bucketEnds := make([]time.Time, len(index))
	for i := range index {
		end := bucketEnd(index, i)
		bucketEnds[i] = end
	}

	for _, event := range events {
		if event.Duration <= 0 {
			continue
		}

		eventEnd := event.StartTime.Add(event.Duration)
		startIndex := sort.Search(len(bucketEnds), func(i int) bool {
			return bucketEnds[i].After(event.StartTime)
		})
		for i := startIndex; i < len(index) && index[i].Before(eventEnd); i++ {
			bucketWidth := bucketEnds[i].Sub(index[i])
			if bucketWidth <= 0 {
				continue
			}

			overlapStart := maxTime(index[i], event.StartTime)
			overlapEnd := minTime(bucketEnds[i], eventEnd)
			overlap := overlapEnd.Sub(overlapStart)
			if overlap > 0 {
				values[i] += event.Intensity * overlap.Seconds() / bucketWidth.Seconds()
			}
		}
	}

	return zipIndexValues(index, values)
}

/**
 * Generates synthetic demand events for a single end-use category over a time interval.
 *
 * Event count is sampled from a Poisson process with the requested mean rate. Start
 * times are sampled uniformly across the interval, while duration and intensity are
 * drawn from simple positive exponential-like distributions centered on the configured means.
 *
 * @param start The inclusive start of the event generation interval.
 * @param end The exclusive end of the event generation interval.
 * @param cfg The configuration controlling event rate, duration, and intensity.
 *
 * @return A sorted slice of positive-intensity demand events fully contained in the requested interval.
 */
func EndUseEvents(start time.Time, end time.Time, cfg EndUseConfig) []DemandEvent {
	if !start.Before(end) || cfg.EventsPerDayMean <= 0 || cfg.DurationMean <= 0 || cfg.IntensityMean <= 0 {
		return nil
	}

	span := end.Sub(start)
	expectedCount := cfg.EventsPerDayMean * span.Hours() / 24.0
	count := poissonCount(expectedCount, cfg.Rand)
	if count == 0 {
		return nil
	}

	events := make([]DemandEvent, 0, count)
	for i := 0; i < count; i++ {
		offset := time.Duration(randomFloat64(cfg.Rand) * float64(span))
		eventStart := start.Add(offset)
		duration := positiveSampleDuration(cfg.DurationMean, cfg.Rand)
		if eventStart.Add(duration).After(end) {
			duration = end.Sub(eventStart)
		}
		if duration <= 0 {
			duration = time.Second
			if eventStart.Add(duration).After(end) {
				eventStart = end.Add(-duration)
			}
		}

		events = append(events, DemandEvent{
			StartTime: eventStart,
			Duration:  duration,
			Intensity: positiveSample(cfg.IntensityMean, cfg.Rand),
		})
	}

	slices.SortFunc(events, func(left DemandEvent, right DemandEvent) int {
		return left.StartTime.Compare(right.StartTime)
	})
	return events
}

/**
 * Generates a stochastic household water-demand series over the provided index.
 *
 * The generator combines several end-use categories with different diurnal
 * activity profiles and occupant-dependent rates. This is still a lightweight
 * synthetic model, but it follows the common end-use decomposition used in the
 * residential water-demand literature more closely than a single undifferentiated
 * pulse stream.
 *
 * @param index The timestamps used for the generated series.
 * @param cfg The household configuration controlling occupant-scaled activity.
 *
 * @return A nonnegative synthetic demand TimeSeries over the provided index.
 */
func HouseholdDemand(index []time.Time, cfg HouseholdDemandConfig) timeseriesgo.TimeSeries {
	if len(index) == 0 || cfg.Occupants <= 0 {
		return timeseriesgo.Empty()
	}

	start := index[0]
	end := index[len(index)-1].Add(inferStep(index))
	occupants := float64(cfg.Occupants)
	sharedScale := 1.0 + 0.65*math.Max(occupants-1, 0)

	events := make([]DemandEvent, 0, int(20*sharedScale))
	specs := []householdEndUseSpec{
		{
			config: EndUseConfig{
				EventsPerDayMean: 0.65 * occupants,
				DurationMean:     7 * time.Minute,
				IntensityMean:    0.12,
				Rand:             cfg.Rand,
			},
			profile: morningEveningProfile(),
		},
		{
			config: EndUseConfig{
				EventsPerDayMean: 4.8 * occupants,
				DurationMean:     75 * time.Second,
				IntensityMean:    0.07,
				Rand:             cfg.Rand,
			},
			profile: allDayProfile(0.8, 1.35, 1.2),
		},
		{
			config: EndUseConfig{
				EventsPerDayMean: 6.5 * occupants,
				DurationMean:     40 * time.Second,
				IntensityMean:    0.03,
				Rand:             cfg.Rand,
			},
			profile: allDayProfile(0.7, 1.15, 1.4),
		},
		{
			config: EndUseConfig{
				EventsPerDayMean: 0.35 * sharedScale,
				DurationMean:     50 * time.Minute,
				IntensityMean:    0.09,
				Rand:             cfg.Rand,
			},
			profile: dayEveningProfile(),
		},
		{
			config: EndUseConfig{
				EventsPerDayMean: 0.55 * sharedScale,
				DurationMean:     55 * time.Minute,
				IntensityMean:    0.05,
				Rand:             cfg.Rand,
			},
			profile: eveningProfile(),
		},
	}

	for _, spec := range specs {
		events = append(events, endUseEventsWithProfile(start, end, spec.config, spec.profile)...)
	}

	return RenderEvents(index, events)
}

type householdEndUseSpec struct {
	config  EndUseConfig
	profile [24]float64
}

func positiveSample(mean float64, rng *rand.Rand) float64 {
	if mean <= 0 {
		return 0
	}

	// Exponential-style draw with the requested mean.
	u := randomFloat64(rng)
	if u == 1 {
		u = 0.999999999
	}
	return -math.Log(1-u) * mean
}

func positiveSampleDuration(mean time.Duration, rng *rand.Rand) time.Duration {
	duration := time.Duration(positiveSample(float64(mean), rng))
	if duration <= 0 {
		return time.Second
	}
	return duration
}

func inferStep(index []time.Time) time.Duration {
	if len(index) < 2 {
		return time.Minute
	}

	step := index[1].Sub(index[0])
	if step <= 0 {
		return time.Minute
	}
	return step
}

func endUseEventsWithProfile(start time.Time, end time.Time, cfg EndUseConfig, profile [24]float64) []DemandEvent {
	if !start.Before(end) || cfg.EventsPerDayMean <= 0 || cfg.DurationMean <= 0 || cfg.IntensityMean <= 0 {
		return nil
	}

	span := end.Sub(start)
	expectedCount := cfg.EventsPerDayMean * span.Hours() / 24.0
	count := poissonCount(expectedCount, cfg.Rand)
	if count == 0 {
		return nil
	}

	events := make([]DemandEvent, 0, count)
	for i := 0; i < count; i++ {
		eventStart := sampleStartByProfile(start, end, profile, cfg.Rand)
		duration := positiveSampleDuration(cfg.DurationMean, cfg.Rand)
		if eventStart.Add(duration).After(end) {
			duration = end.Sub(eventStart)
		}
		if duration <= 0 {
			duration = time.Second
			if eventStart.Add(duration).After(end) {
				eventStart = end.Add(-duration)
			}
		}

		events = append(events, DemandEvent{
			StartTime: eventStart,
			Duration:  duration,
			Intensity: positiveSample(cfg.IntensityMean, cfg.Rand),
		})
	}

	slices.SortFunc(events, func(left DemandEvent, right DemandEvent) int {
		return left.StartTime.Compare(right.StartTime)
	})
	return events
}

func sampleStartByProfile(start time.Time, end time.Time, profile [24]float64, rng *rand.Rand) time.Time {
	type bucket struct {
		start  time.Time
		end    time.Time
		weight float64
	}

	cursor := start.Truncate(time.Hour)
	if cursor.After(start) {
		cursor = cursor.Add(-time.Hour)
	}

	buckets := make([]bucket, 0)
	totalWeight := 0.0
	for bucketStart := cursor; bucketStart.Before(end); bucketStart = bucketStart.Add(time.Hour) {
		bucketEnd := bucketStart.Add(time.Hour)
		activeStart := bucketStart
		if activeStart.Before(start) {
			activeStart = start
		}
		activeEnd := bucketEnd
		if activeEnd.After(end) {
			activeEnd = end
		}
		if !activeStart.Before(activeEnd) {
			continue
		}

		weight := profile[bucketStart.Hour()] * activeEnd.Sub(activeStart).Hours()
		if weight <= 0 {
			continue
		}

		buckets = append(buckets, bucket{
			start:  activeStart,
			end:    activeEnd,
			weight: weight,
		})
		totalWeight += weight
	}

	if len(buckets) == 0 || totalWeight <= 0 {
		return start.Add(time.Duration(randomFloat64(rng) * float64(end.Sub(start))))
	}

	target := randomFloat64(rng) * totalWeight
	cumulative := 0.0
	for _, candidate := range buckets {
		cumulative += candidate.weight
		if target <= cumulative {
			return candidate.start.Add(time.Duration(randomFloat64(rng) * float64(candidate.end.Sub(candidate.start))))
		}
	}

	last := buckets[len(buckets)-1]
	return last.start.Add(time.Duration(randomFloat64(rng) * float64(last.end.Sub(last.start))))
}

func morningEveningProfile() [24]float64 {
	var profile [24]float64
	for hour := range profile {
		profile[hour] = 0.15
	}
	for _, hour := range []int{6, 7, 8} {
		profile[hour] = 1.8
	}
	for _, hour := range []int{19, 20, 21} {
		profile[hour] = 1.4
	}
	return profile
}

func eveningProfile() [24]float64 {
	var profile [24]float64
	for hour := range profile {
		profile[hour] = 0.1
	}
	for _, hour := range []int{18, 19, 20, 21} {
		profile[hour] = 1.9
	}
	for _, hour := range []int{12, 13} {
		profile[hour] = 0.7
	}
	return profile
}

func dayEveningProfile() [24]float64 {
	var profile [24]float64
	for hour := range profile {
		profile[hour] = 0.1
	}
	for _, hour := range []int{10, 11, 12, 13, 14, 15, 19, 20} {
		profile[hour] = 1.2
	}
	return profile
}

func allDayProfile(night float64, morning float64, evening float64) [24]float64 {
	var profile [24]float64
	for hour := range profile {
		switch {
		case hour >= 0 && hour < 5:
			profile[hour] = night
		case hour >= 5 && hour < 11:
			profile[hour] = morning
		case hour >= 11 && hour < 17:
			profile[hour] = 1.0
		default:
			profile[hour] = evening
		}
	}
	return profile
}

func bucketEnd(index []time.Time, position int) time.Time {
	if position < len(index)-1 && index[position+1].After(index[position]) {
		return index[position+1]
	}
	return index[position].Add(inferStep(index))
}

func minTime(left time.Time, right time.Time) time.Time {
	if left.Before(right) {
		return left
	}
	return right
}

func maxTime(left time.Time, right time.Time) time.Time {
	if left.After(right) {
		return left
	}
	return right
}
