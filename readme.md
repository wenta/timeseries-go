# timeseries-go

A Go library for working with time series, from raw points to analysis-ready outputs.


## Overview

`timeseries-go` provides:

- core `TimeSeries` and aligned-series primitives
- transforms such as slicing, reindexing, resampling, interpolation, fill, and cumulative operations
- statistical helpers including rolling windows, covariance, correlation, and ACF
- forecasting with naive, exponential smoothing, Holt, and Holt-Winters methods
- decomposition with classical additive decomposition, STL, and MSTL
- anomaly detection helpers
- synthetic data generators
- plotting to HTML, SVG, and PNG

## Installation

```bash
go get github.com/wenta/timeseries-go
```

## Minimal Example

```go
package main

import (
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/plot"
	"github.com/wenta/timeseries-go/stats"
)

func main() {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	ts := timeseriesgo.FromDataPoints([]timeseriesgo.DataPoint{
		{Timestamp: base, Value: 100},
		{Timestamp: base.Add(24 * time.Hour), Value: 120},
		{Timestamp: base.Add(48 * time.Hour), Value: 115},
		{Timestamp: base.Add(72 * time.Hour), Value: 130},
	})

	smoothed := stats.MovingAverage(ts, 48*time.Hour)

	_ = plot.Save("out/moving-average.html", []plot.Series{
		{Label: "raw", Data: ts, Color: plot.Gold},
		{Label: "moving average", Data: smoothed, Color: plot.DeepSkyBlue},
	})
}
```

## Package Guide

### `timeseriesgo`

Use this package to construct series and perform core operations:

- `Empty`, `EmptyLabeled`, `FromDataPoints`, `Zip`
- `Slice`, `MapValues`, `Filter`
- `Resample`, `Interpolate`, `Step`
- `Join`, `JoinLeft`, `JoinOuter`, `Merge`
- `ForwardFill`, `BackwardFill`, `FillMissing`
- `Reindex`, `ReindexNearest`

### `stats`

Use this package for descriptive and rolling statistics:

- `GetMeanAndVariance`
- `Covariance`
- `Correlation`
- `ACF`
- `MovingAverage`
- `RollingSum`, `RollingMean`, `RollingMin`, `RollingMax`, `RollingStdDev`, `RollingMedian`
- `ZNormalize`, `EMA`, `EWVariance`

### `forecast`

Use this package for baseline forecasting methods:

- `Naive`
- `SimpleExponentialSmoothing`
- `DoubleExponentialSmoothing`
- `DoubleExponentialSmoothingEstimated`
- `TripleExponentialSmoothing`

### `decompose`

Use this package to split a series into components:

- `SeasonalDecompose`
- `STL`
- `MSTL`

### `anomaly`

Use this package to score or flag anomalous points:

- `ZScore`
- `RobustZScore`
- `FindAnomaliesWithZScore`
- `FindAnomaliesWithRobustZScore`
- `FindSpikeAnomalies`
- `FindDropAnomalies`
- `FindFlatlineAnomalies`

### `generator`

Use this package to build synthetic time-series datasets:

- deterministic generators
- noise generators
- block bootstrap and seasonal resampling
- event rendering
- household demand generation

### `plot`

Use this package to render series to files:

- `Save`
- `SaveHTML`
- `SaveSVG`
- `SavePNG`
- `HTML`
- `SVG`
- `PNG`

## Examples

Generate all example reports:

```bash
go run ./examples/all
```

Useful entry points:

```bash
go run ./examples
go run ./examples/forecasting
go run ./examples/decomposition
go run ./examples/anomalies
go run ./examples/generators
```

Resources:

- example dataset: [examples/data/air_passengers.csv](examples/data/air_passengers.csv)
- generated reports: [examples/out/index.html](examples/out/index.html)
- package docs: https://pkg.go.dev/github.com/wenta/timeseries-go
