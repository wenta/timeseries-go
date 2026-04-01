package generator

import (
	"math/rand/v2"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

/**
 * Resamples a TimeSeries by drawing values from contiguous non-overlapping source blocks.
 *
 * @param ts The source TimeSeries providing values to sample from.
 * @param outputIndex The timestamps to use for the generated output series.
 * @param blockSize The number of consecutive source values per sampled block.
 *
 * @return A TimeSeries with timestamps from outputIndex and values copied from sampled source blocks.
 */
func BlockBootstrap(ts timeseriesgo.TimeSeries, outputIndex []time.Time, blockSize int) timeseriesgo.TimeSeries {
	return bootstrap(ts, outputIndex, blockSize, false)
}

/**
 * Resamples a TimeSeries by drawing values from contiguous overlapping source blocks.
 *
 * @param ts The source TimeSeries providing values to sample from.
 * @param outputIndex The timestamps to use for the generated output series.
 * @param blockSize The number of consecutive source values per sampled block.
 *
 * @return A TimeSeries with timestamps from outputIndex and values copied from sampled source blocks.
 */
func MovingBlockBootstrap(ts timeseriesgo.TimeSeries, outputIndex []time.Time, blockSize int) timeseriesgo.TimeSeries {
	return bootstrap(ts, outputIndex, blockSize, true)
}

/**
 * Resamples source values by seasonal bucket key while preserving output timestamps.
 *
 * For each timestamp in outputIndex, keyFn determines the seasonal bucket. If the
 * source series does not contain that bucket, the function falls back to sampling
 * from all source values.
 *
 * @param ts The source TimeSeries providing seasonal buckets and values.
 * @param outputIndex The timestamps to use for the generated output series.
 * @param keyFn A function mapping timestamps to seasonal bucket keys such as hour or weekday.
 *
 * @return A TimeSeries with timestamps from outputIndex and values sampled from matching seasonal buckets or the global fallback pool.
 */
func SeasonalResample(ts timeseriesgo.TimeSeries, outputIndex []time.Time, keyFn func(time.Time) string) timeseriesgo.TimeSeries {
	if ts.IsEmpty() || len(outputIndex) == 0 || keyFn == nil {
		return timeseriesgo.Empty()
	}

	points := ts.DataPoints()
	byKey := make(map[string][]float64)
	allValues := make([]float64, 0, len(points))
	for _, point := range points {
		key := keyFn(point.Timestamp)
		byKey[key] = append(byKey[key], point.Value)
		allValues = append(allValues, point.Value)
	}

	values := make([]float64, len(outputIndex))
	for i, timestamp := range outputIndex {
		candidates := byKey[keyFn(timestamp)]
		if len(candidates) == 0 {
			candidates = allValues
		}
		values[i] = candidates[rand.IntN(len(candidates))]
	}

	return zipIndexValues(outputIndex, values)
}

func bootstrap(ts timeseriesgo.TimeSeries, outputIndex []time.Time, blockSize int, moving bool) timeseriesgo.TimeSeries {
	if ts.IsEmpty() || len(outputIndex) == 0 || blockSize <= 0 {
		return timeseriesgo.Empty()
	}

	sourceValues := ts.Values()
	starts := bootstrapBlockStarts(len(sourceValues), blockSize, moving)
	if len(starts) == 0 {
		return timeseriesgo.Empty()
	}

	values := make([]float64, 0, len(outputIndex))
	for len(values) < len(outputIndex) {
		start := starts[rand.IntN(len(starts))]
		end := minInt(start+blockSize, len(sourceValues))
		values = append(values, sourceValues[start:end]...)
	}

	values = values[:len(outputIndex)]
	return zipIndexValues(outputIndex, values)
}

func bootstrapBlockStarts(length int, blockSize int, moving bool) []int {
	if length <= 0 || blockSize <= 0 {
		return nil
	}

	if blockSize >= length {
		return []int{0}
	}

	starts := make([]int, 0)
	if moving {
		for start := 0; start < length; start++ {
			starts = append(starts, start)
		}
		return starts
	}

	for start := 0; start < length; start += blockSize {
		starts = append(starts, start)
	}
	return starts
}
