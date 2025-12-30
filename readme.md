# Time Series library

Library for processing time series in Go.

Package docs: https://pkg.go.dev/github.com/wenta/timeseries-go

Planned work: see [todo.md](todo.md).


## Common setup

```go
package main

import (
	"encoding/csv"
	"strings"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/anomaly"
	"github.com/wenta/timeseries-go/forecast"
	"github.com/wenta/timeseries-go/generator"
	"github.com/wenta/timeseries-go/metrics"
	"github.com/wenta/timeseries-go/stats"
	"github.com/wenta/timeseries-go/tsio"
)

func main() {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	ts := timeseriesgo.Empty()
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 10})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(time.Hour), Value: 12})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: base.Add(2 * time.Hour), Value: 9})
}
```

## Implemented functions

#### Core construction and access (timeseriesgo)
Create and inspect series basics.
```go
labeled := timeseriesgo.EmptyLabeled("cpu")
labeled.AddPoint(timeseriesgo.DataPoint{Timestamp: base, Value: 3})
labeled.Length()

points := []timeseriesgo.DataPoint{
	{Timestamp: base.Add(3 * time.Hour), Value: 7},
}
fromPoints := timeseriesgo.FromDataPoints(points)
fromPoints.Print()

timestamps := []time.Time{base, base.Add(time.Hour)}
values := []float64{10, 11}
zipped, _ := timeseriesgo.Zip(timestamps, values)

tsTimes, tsValues := zipped.UnZip()

vals := ts.Values()
times := ts.Timestamps()
raw := ts.DataPoints()

first, _ := ts.Head()
last, _ := ts.Last()
tail := ts.Tail()

resolution, _ := ts.Resolution()

ts.Print()
```

#### Slicing and transforms (timeseriesgo)
Slice, map, and filter values.
```go
start := base.Add(30 * time.Minute)
end := base.Add(2 * time.Hour)

sub := ts.Slice(start, end)
scaled := ts.MapValues(func(v float64) float64 { return v * 2 })
shifted := ts.Map(func(dp timeseriesgo.DataPoint) timeseriesgo.DataPoint {
	dp.Value += 1
	return dp
})
high := ts.Filter(func(dp timeseriesgo.DataPoint) bool { return dp.Value > 10 })
```

#### Resampling and interpolation (timeseriesgo)
Resample on a fixed grid.
```go
rs := ts.Resample(time.Minute, func(a, b timeseriesgo.DataPoint, t time.Time) float64 {
	return a.Value
})
rsDefault := ts.ResampleWithDefaultValue(time.Minute, 0)
lin := ts.Interpolate(time.Minute)
stepSeries := ts.Step(time.Minute)
```

#### Grouping and rolling (timeseriesgo, stats)
Aggregate by time buckets and compute rolling stats.
```go
hourly := ts.GroupByTime(
	func(t time.Time) time.Time { return t.Truncate(time.Hour) },
	func(points []timeseriesgo.DataPoint) float64 { return float64(len(points)) },
)
roll := ts.RollingWindow(time.Hour, func(values []float64) float64 {
	return values[len(values)-1]
})
ma := stats.MovingAverage(ts, time.Hour)

```

#### Joins and merge (timeseriesgo)
Combine multiple series.
```go
other := ts.MapValues(func(v float64) float64 { return v - 1 })

merged := ts.Merge(other)
inner := ts.Join(other)
leftJoin := ts.JoinLeft(other, 0)
outer := ts.JoinOuter(other, 0, 0)
```

#### Aligned series helpers (timeseriesgo)
Work with joined series.
```go
other := ts.MapValues(func(v float64) float64 { return v + 1 })

aligned := ts.Join(other)
count := aligned.Length()
pairs := aligned.DataPoints()
pairDiff := aligned.MapValuesWithReduce(func(l, r float64) float64 { return l - r })

aligned.Print()
```

#### Statistics (timeseriesgo, stats)
Basic stats and transforms.
```go
min, _ := ts.Min()
max, _ := ts.Max()
total := ts.Sum()
p95, _ := ts.Percentile(95)
median, _ := ts.Median()
diffSeries := ts.Differentiate()
integ := ts.Integrate()
mv, _ := stats.GetMeanAndVariance(ts)
```

#### Metrics (metrics)
Compare series.
```go
other := ts.MapValues(func(v float64) float64 { return v - 1 })

mse, _ := metrics.MSE(ts, other)
rmse, _ := metrics.RMSE(ts, other)
mae, _ := metrics.MAE(ts, other)
mad, _ := metrics.MAD(ts)

```

#### Forecasting (forecast)
Naive forecasts.
```go
fc := forecast.Naive(ts, 3)
```

#### Generators (generator)
Create synthetic series.
```go
index := generator.MakeSeriesIndex(base, time.Hour, 4)
constant := generator.Constant(index, 5)
walk := generator.RandomWalk(index, 10)

patternIndex := generator.MakeSeriesIndex(base, time.Hour, 2)
pattern := generator.Constant(patternIndex, 1)
loop := generator.Repeat(pattern, base, base.Add(4*time.Hour))
```

#### Anomaly detection (anomaly)
Detect spikes and anomalies.
```go
zs, _ := anomaly.ZScore(ts)
flags, _ := anomaly.FindAnomaliesWithZScore(ts)
rz, _ := anomaly.RobustZScore(ts)
rflags, _ := anomaly.FindAnomaliesWithRobustZScore(ts)
spikes, _ := anomaly.FindSpikeAnomalies(ts, 3)
drops, _ := anomaly.FindDropAnomalies(ts, 3)
flat, _ := anomaly.FindFlatlineAnomalies(ts, 0.1, 2)
```

#### IO (tsio)
CSV serialization.
```go
csvStr, _ := tsio.ToString(ts)
csvStr2, _ := tsio.ToStringWithTimeFormat(ts, time.RFC3339)

r := csv.NewReader(strings.NewReader(csvStr))
parsed, _ := tsio.FromString(*r, "cpu")

r2 := csv.NewReader(strings.NewReader(csvStr))
parsed2, _ := tsio.FromStringWithTimeFormat(*r2, time.RFC3339, "cpu")

```

# Join in!

We are happy to receive bug reports, fixes, documentation enhancements,
and other improvements.

Please report bugs via the
[github issue tracker](http://github.com/wenta/timeseries-go/issues).
