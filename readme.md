# Time Series library

Library for processing time series in Go.

Package docs: https://pkg.go.dev/github.com/wenta/timeseries-go

Planned work: see [todo.md](todo.md).

## Example programs

```bash
go run ./examples
go run ./examples/plotting
go run ./examples/forecasting
go run ./examples/anomalies
go run ./examples/decomposition
go run ./examples/transformations
go run ./examples/generators
```

Shared example dataset: [examples/data/air_passengers.csv](examples/data/air_passengers.csv)


## Common setup

```go
package main

import (
	"encoding/csv"
	"strings"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/anomaly"
	"github.com/wenta/timeseries-go/decompose"
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

#### Decomposition (decompose)
Additive seasonal decomposition, STL, and MSTL for multiple seasonal periods.
```go
result, _ := decompose.SeasonalDecompose(ts, 12)
stl, _ := decompose.STL(ts, decompose.STLConfig{Period: 12, Robust: true})
mstl, _ := decompose.MSTL(ts, decompose.MSTLConfig{Periods: []int{24, 24 * 7}})

trend := result.Trend
seasonal := result.Seasonal
residual := result.Residual

stlTrend := stl.Trend
stlSeasonal := stl.Seasonal
stlResidual := stl.Residual

mstlTrend := mstl.Trend
mstlDaily := mstl.Seasonal[0].Series
mstlWeekly := mstl.Seasonal[1].Series
mstlResidual := mstl.Residual
```

`SeasonalDecompose` implements the classical additive decomposition approach based on centered moving averages.
`STL` follows the method introduced by Cleveland, Cleveland, McRae, and Terpenning, "STL: A Seasonal-Trend Decomposition Procedure Based on Loess" (1990).
`MSTL` extends STL to multiple seasonal periods by iteratively estimating one seasonal component per period.

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
Naive forecasts, simple exponential smoothing, Holt linear trend, and Holt-Winters additive seasonal forecasting.
```go
fc := forecast.Naive(ts, 3)
ses := forecast.SimpleExponentialSmoothing(ts, 0.2, 3)
holt := forecast.DoubleExponentialSmoothing(ts, 0.8, 0.2, 3)
holtEstimated := forecast.DoubleExponentialSmoothingEstimated(ts, 0.8, 0.2, 3)
holtWinters := forecast.TripleExponentialSmoothing(ts, 0.6, 0.3, 0.2, 12, 3)
```

#### Plotting (plot)
Generate browser-friendly HTML charts or static SVG/PNG files.
```go
comparison := []plot.Series{
	{Label: "actual", Data: ts, Color: plot.Gold},
	{Label: "forecast", Data: forecastTS, Color: plot.DeepSkyBlue, Style: plot.LinePoints},
}

_ = plot.Save("out/actual-vs-forecast.html", comparison, plot.Title("Actual vs Forecast"))
_ = plot.SaveSVG("out/actual-vs-forecast.svg", comparison, plot.Title("Actual vs Forecast"))
_ = plot.SavePNG("out/actual-vs-forecast.png", comparison, plot.Title("Actual vs Forecast"))

htmlBytes, _ := plot.HTML(comparison, plot.Title("Actual vs Forecast"))
svgBytes, _ := plot.SVG(comparison, plot.Title("Actual vs Forecast"))
pngBytes, _ := plot.PNG(comparison, plot.Title("Actual vs Forecast"))
```

#### Generators (generator)
Create synthetic series.
```go
index := generator.MakeSeriesIndex(base, time.Hour, 4)
constant := generator.Constant(index, 5)
walk := generator.RandomWalk(index, 10)
noise := generator.RandomNoise(index, 0, 1)

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
