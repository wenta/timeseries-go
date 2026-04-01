# timeseries-go — Synthetic Time Series Roadmap for Water-in-Pipes Use Cases

This document defines a focused set of GitHub issues for extending `timeseries-go` with synthetic time series generation capabilities that are useful for water demand, sewer flow, and related pipe-network simulations.

The goal is **not** to turn the library into a full hydraulic solver. The goal is to add reusable, well-scoped time-series building blocks that fit the current structure and style of the repository.

## Design intent

The new work should follow the current repository style:

- small, focused functions
- clear package placement
- simple public APIs
- deterministic and readable tests
- features added incrementally rather than through large framework refactors

## Scope

This roadmap covers two stages.

### Stage 1
General-purpose generators and resampling helpers.

### Stage 2
Domain-oriented synthetic demand and sewer-flow helpers built on top of Stage 1.

---

# Stage 1 — General generators and resampling helpers

## Issue 1 — Add uniform noise time series generator

**Title**

`Add uniform noise time series generator`

**Description**

Implement a generator function to create uniform random noise for a given index and value range.

This should be a natural extension of the existing generator package, next to `RandomNoise`, `RandomWalk`, `Constant`, and `Repeat`.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
func UniformNoise(index []time.Time, min float64, max float64) timeseriesgo.TimeSeries
```

**Behavior expectations**

- use the provided timestamps unchanged
- generate one value per timestamp
- values should lie in the closed or practical floating-point range `[min, max]`
- if `min == max`, all values should equal that constant
- if the input index is empty, return an empty series

**Testing expectations**

- empty index returns empty series
- `min == max` returns a constant-like series
- generated timestamps match the input index exactly
- values stay within the expected bounds

---

## Issue 2 — Add pulse train time series generator

**Title**

`Add pulse train time series generator`

**Description**

Add a generator function for synthetic pulse-based series. This should be useful for short bursts, device activity, and simple demand events.

This is an important primitive for later synthetic water-demand work.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
type Pulse struct {
	StartIndex int
	Duration   int
	Amplitude  float64
}

func PulseTrain(index []time.Time, pulses []Pulse) timeseriesgo.TimeSeries
```

**Behavior expectations**

- create a zero-baseline series on the provided index
- each pulse adds `Amplitude` for `Duration` points starting at `StartIndex`
- overlapping pulses should add together
- pulses partially outside the index should be clipped safely
- empty input index should return an empty series

**Testing expectations**

- single pulse
- multiple pulses
- overlapping pulses
- out-of-range pulse clipping
- empty index

---

## Issue 3 — Add Poisson event index generator

**Title**

`Add Poisson event index generator`

**Description**

Add a helper that generates event indices using a Poisson-style process. This will be useful for building more advanced stochastic generators later.

This helper does not need to return a `TimeSeries`; returning integer indices is enough.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
func PoissonEventIndices(length int, lambda float64) []int
```

**Behavior expectations**

- return indices in ascending order
- all returned indices must be valid for `[0, length)`
- `lambda <= 0` should return no events
- output may contain zero events

**Testing expectations**

- `length == 0`
- `lambda == 0`
- all returned indices in range
- output is sorted

---

## Issue 4 — Add block bootstrap resampler

**Title**

`Add block bootstrap resampler`

**Description**

Implement a block bootstrap helper for time series resampling. The output series should preserve timestamps from the provided output index while sampling value blocks from the source series.

This is one of the most useful ways to generate synthetic series while retaining short-range autocorrelation.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
func BlockBootstrap(ts timeseriesgo.TimeSeries, outputIndex []time.Time, blockSize int) timeseriesgo.TimeSeries
```

**Behavior expectations**

- preserve timestamps from `outputIndex`
- sample contiguous blocks of values from the input series
- concatenate sampled blocks until the output length is reached
- if input is empty or `outputIndex` is empty, return an empty series
- if `blockSize <= 0`, return an empty series

**Testing expectations**

- empty source series
- empty output index
- invalid block size
- correct output length
- output timestamps equal `outputIndex`
- output values come from input series values

---

## Issue 5 — Add moving block bootstrap resampler

**Title**

`Add moving block bootstrap resampler`

**Description**

Extend bootstrap support with a moving block bootstrap variant, where blocks can start at any valid point in the source series.

This should be a separate function rather than an extra mode flag, to stay aligned with the repository’s current style of small, explicit APIs.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
func MovingBlockBootstrap(ts timeseriesgo.TimeSeries, outputIndex []time.Time, blockSize int) timeseriesgo.TimeSeries
```

**Behavior expectations**

- preserve timestamps from `outputIndex`
- sample overlapping candidate blocks from the input series
- concatenate sampled blocks until output is full
- handle short input safely

**Testing expectations**

- output length matches `len(outputIndex)`
- output timestamps equal `outputIndex`
- invalid inputs return empty series

---

## Issue 6 — Add seasonal resampler by time key

**Title**

`Add seasonal resampler by time key`

**Description**

Add a resampling helper that draws values only from points belonging to the same seasonal group, for example hour of day or weekday.

This should make it possible to preserve simple diurnal or weekday structure during synthetic generation.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
func SeasonalResample(ts timeseriesgo.TimeSeries, outputIndex []time.Time, keyFn func(time.Time) string) timeseriesgo.TimeSeries
```

**Behavior expectations**

- for each timestamp in `outputIndex`, compute its seasonal key using `keyFn`
- draw a value from source points having the same key
- preserve timestamps from `outputIndex`
- if no matching seasonal bucket exists for a key, use a simple fallback strategy
- fallback strategy should be documented in code comments and tests

**Testing expectations**

- hourly grouping example
- weekday grouping example
- fallback behavior when a bucket is missing

---

## Issue 7 — Add k-nearest-neighbor resampler for univariate series

**Title**

`Add k-nearest-neighbor resampler for univariate series`

**Description**

Implement a simple k-nearest-neighbor resampler for univariate time series using a fixed history window.

This is a higher-value feature than plain random sampling because it attempts to continue the series from locally similar historical states.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
func ResampleKNN(ts timeseriesgo.TimeSeries, outputIndex []time.Time, window int, k int) timeseriesgo.TimeSeries
```

**Behavior expectations**

- preserve timestamps from `outputIndex`
- use a trailing window of recent generated values or source context
- find nearest historical windows in the source series
- sample the next value from one of the `k` nearest matches
- if the input is too short for the requested window, return an empty series
- if `window <= 0` or `k <= 0`, return an empty series

**Testing expectations**

- invalid parameters
- too-short source series
- correct output length
- timestamps preserved

**Implementation note**

A simple Euclidean distance over raw values is enough for the first version.

---

## Issue 8 — Add lag autocorrelation helper

**Title**

`Add lag autocorrelation helper`

**Description**

Add an autocorrelation helper for a single time series. This will be useful for evaluating synthetic generators and bootstrap methods.

This fits naturally next to other statistics helpers already present in the repository.

**Implementation location**

- `stats/stats.go`
- tests in `stats/stats_test.go`

**Suggested API**

```go
func Autocorrelation(ts timeseriesgo.TimeSeries, lag int) (float64, error)
```

**Behavior expectations**

- compute lag autocorrelation on aligned positions of one series
- return an error for empty series
- return an error for invalid lag values
- return an error when variance is zero

**Testing expectations**

- empty series
- invalid lag
- zero-variance series
- a simple hand-checkable example

---

## Issue 9 — Add series comparison helper for generator evaluation

**Title**

`Add series comparison helper for generator evaluation`

**Description**

Add a helper for comparing original and synthetic series using a few basic statistics such as mean, variance, and selected lag autocorrelations.

This is intended to support development and validation of the new generator features.

**Implementation location**

- `stats/stats.go` or `metrics/metrics.go`
- tests in the matching test file

**Suggested API**

```go
type SeriesComparison struct {
	MeanOriginal      float64
	MeanSynthetic     float64
	VarianceOriginal  float64
	VarianceSynthetic float64
}

func CompareSeriesStats(original, synthetic timeseriesgo.TimeSeries) (SeriesComparison, error)
```

**Behavior expectations**

- compute a small, stable comparison summary
- return an error if either series is empty
- document whether comparison is done on raw values or aligned timestamps

**Testing expectations**

- empty-series error case
- identical series
- clearly different series

---

# Stage 2 — Water-demand and sewer-flow helpers

## Issue 10 — Add event rendering helper for synthetic demand series

**Title**

`Add event rendering helper for synthetic demand series`

**Description**

Add a helper that converts a list of demand events into a regular time series on a fixed index.

This is the bridge between event-based stochastic models and regular time-step series.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
type DemandEvent struct {
	StartTime time.Time
	Duration  time.Duration
	Intensity float64
}

func RenderEvents(index []time.Time, events []DemandEvent) timeseriesgo.TimeSeries
```

**Behavior expectations**

- create a zero-baseline series over `index`
- each event contributes `Intensity` over its active duration
- overlapping events should add together
- timestamps are exactly those from `index`
- empty index returns empty series

**Testing expectations**

- single event
- overlapping events
- event shorter than one time step
- empty index

---

## Issue 11 — Add simple household demand generator

**Title**

`Add simple household demand generator`

**Description**

Implement a simple stochastic household water demand generator producing sparse pulse-like usage over time.

The first version should stay simple and generate realistic-looking time series rather than attempt a full end-use research model.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
type HouseholdDemandConfig struct {
	Occupants int
}

func HouseholdDemand(index []time.Time, cfg HouseholdDemandConfig) timeseriesgo.TimeSeries
```

**Behavior expectations**

- use the provided index as the time axis
- produce sparse nonnegative usage
- higher `Occupants` should generally produce more total demand
- empty index should return empty series
- invalid occupant count should be handled safely

**Testing expectations**

- empty index
- one occupant versus multiple occupants
- all values are nonnegative
- timestamps preserved

**Implementation note**

A simple composition of random events and pulses is enough for the first version.

---

## Issue 12 — Add end-use demand event generator

**Title**

`Add end-use demand event generator`

**Description**

Add a helper for generating synthetic demand events for a single end-use category such as shower, toilet, or faucet.

This should support later composition into household demand.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
type EndUseConfig struct {
	EventsPerDayMean float64
	DurationMean     time.Duration
	IntensityMean    float64
}

func EndUseEvents(start time.Time, end time.Time, cfg EndUseConfig) []DemandEvent
```

**Behavior expectations**

- generate events between `start` and `end`
- event count should roughly reflect `EventsPerDayMean`
- event durations and intensities should be positive
- returned events should be sorted by start time

**Testing expectations**

- empty or invalid interval
- sorted output
- all events lie within the requested time range

---

## Issue 13 — Add aggregation helper for multiple time series

**Title**

`Add aggregation helper for multiple time series`

**Description**

Add a helper that sums multiple time series with matching timestamps. This will be useful for aggregating synthetic household demand to node level.

This seems useful beyond the water-demand use case and should likely live in the root package rather than only inside `generator`.

**Implementation location**

- `timeseries.go` or a nearby helper file in the root package
- tests in `timeseries_test.go`

**Suggested API**

```go
func SumSeries(seriesList []TimeSeries) (TimeSeries, error)
```

**Behavior expectations**

- sum aligned values point by point
- return an error when timestamps do not match
- return an empty series for an empty input list or clearly document a different choice

**Testing expectations**

- identical timestamp alignment
- mismatch error case
- empty list behavior

---

## Issue 14 — Add weighted aggregation helper for multiple time series

**Title**

`Add weighted aggregation helper for multiple time series`

**Description**

Add a weighted aggregation helper for multiple aligned series.

This should complement `SumSeries` and support simple allocation or mixing use cases.

**Implementation location**

- near `SumSeries`
- tests in `timeseries_test.go`

**Suggested API**

```go
func WeightedSumSeries(seriesList []TimeSeries, weights []float64) (TimeSeries, error)
```

**Behavior expectations**

- require matching timestamps across all series
- require `len(weights) == len(seriesList)`
- return clear errors for invalid inputs

**Testing expectations**

- happy path with two aligned series
- weight count mismatch
- timestamp mismatch

---

## Issue 15 — Add simple sewer flow converter from demand series

**Title**

`Add simple sewer flow converter from demand series`

**Description**

Implement a helper that converts water demand into delayed and scaled sewer flow. This should support a return factor and fixed delay.

The first version should stay intentionally simple.

**Implementation location**

- `generator/generators.go` or a new package if the implementation feels too domain-specific
- tests in the matching test file

**Suggested API**

```go
type SewerConversionConfig struct {
	ReturnFactor float64
	Delay        time.Duration
}

func DemandToSewerFlow(ts timeseriesgo.TimeSeries, cfg SewerConversionConfig) timeseriesgo.TimeSeries
```

**Behavior expectations**

- preserve the same regular time-step structure as the input series
- scale flow by `ReturnFactor`
- shift the signal by `Delay`
- values should remain nonnegative if the input is nonnegative
- empty input should return empty series

**Testing expectations**

- empty input
- zero delay
- positive delay
- simple mass/scale sanity check

---

## Issue 16 — Add leak injection helper

**Title**

`Add leak injection helper`

**Description**

Add a helper for injecting a persistent leak into an existing time series from a selected start time.

This should be useful both for synthetic scenario creation and anomaly-detection examples.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
func InjectLeak(ts timeseriesgo.TimeSeries, start time.Time, magnitude float64) timeseriesgo.TimeSeries
```

**Behavior expectations**

- preserve all timestamps
- add `magnitude` from `start` onward
- leave earlier points unchanged
- return a new series rather than mutating input behavior unexpectedly

**Testing expectations**

- leak starting mid-series
- leak starting before first timestamp
- leak starting after last timestamp

---

## Issue 17 — Add burst injection helper

**Title**

`Add burst injection helper`

**Description**

Add a helper for injecting a short burst anomaly into an existing time series.

This is useful for demand spikes, pipe bursts, and anomaly-detection examples.

**Implementation location**

- `generator/generators.go`
- tests in `generator/generators_test.go`

**Suggested API**

```go
func InjectBurst(ts timeseriesgo.TimeSeries, start time.Time, duration time.Duration, magnitude float64) timeseriesgo.TimeSeries
```

**Behavior expectations**

- preserve timestamps
- add `magnitude` only over the requested burst interval
- leave points outside the interval unchanged
- return a new series

**Testing expectations**

- burst fully inside series range
- burst partially overlapping range
- burst outside range

---

# Recommended implementation order

Recommended order for opening or implementing the issues:

1. Add uniform noise time series generator
2. Add pulse train time series generator
3. Add block bootstrap resampler
4. Add k-nearest-neighbor resampler for univariate series
5. Add lag autocorrelation helper
6. Add event rendering helper for synthetic demand series
7. Add simple household demand generator
8. Add aggregation helper for multiple time series
9. Add simple sewer flow converter from demand series
10. Add leak injection helper
11. Add burst injection helper

This order gives the repository useful generic features first, then adds water-demand functionality on top of them.

---

# What success looks like

After these issues are implemented, `timeseries-go` should be able to do the following:

- generate basic synthetic series such as uniform noise and pulse trains
- resample historical series using bootstrap and simple kNN-based methods
- evaluate synthetic output with simple statistical checks
- render event-based demand into a regular time-step series
- build simple synthetic household demand profiles
- aggregate multiple synthetic demand series
- derive simple sewer-flow series from synthetic demand
- inject leak and burst anomalies into otherwise normal synthetic data

That would make the library significantly more useful for:

- synthetic data generation
- testing anomaly detectors
- demo pipelines
- forecasting experiments
- water-demand and sewer-flow prototyping

---

# Notes for Codex

When implementing these issues, keep the repository style consistent:

- prefer small, direct functions over large abstractions
- keep the API explicit and easy to read
- extend the existing `generator` package where possible
- add focused tests next to the implementation
- follow the existing naming style already used in the repository
- do not introduce a large framework or advanced dependency graph for the first iteration

The first goal is a clean, useful, incremental extension of the library.
