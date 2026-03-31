package decompose

import (
	"errors"
	"math"

	timeseriesgo "github.com/wenta/timeseries-go"
)

// DecomposeResult contains the additive components produced by SeasonalDecompose.
type DecomposeResult struct {
	Trend    timeseriesgo.TimeSeries
	Seasonal timeseriesgo.TimeSeries
	Residual timeseriesgo.TimeSeries
}

/**
 * Decomposes a TimeSeries into additive trend, seasonal, and residual components.
 *
 * This implements the classical additive decomposition approach based on
 * centered moving averages: trend is estimated with a centered moving average,
 * seasonal factors are estimated from the de-trended series, and residual is
 * the remaining component.
 *
 * @param ts The TimeSeries to decompose. Expected that ts is already sorted by timestamp.
 * @param period The number of observations in a full seasonal cycle.
 *
 * @return A DecomposeResult containing full-length Trend, Seasonal, and Residual series.
 * @return An error if the series is empty, the period is invalid, or fewer than two full cycles are available.
 */
func SeasonalDecompose(ts timeseriesgo.TimeSeries, period int) (DecomposeResult, error) {
	if ts.IsEmpty() {
		return DecomposeResult{}, errors.New("TimeSeries is empty")
	}
	if period < 2 {
		return DecomposeResult{}, errors.New("period must be at least 2")
	}

	points := ts.DataPoints()
	if len(points) < 2*period {
		return DecomposeResult{}, errors.New("seasonal decomposition requires at least two complete cycles")
	}

	kernel := seasonalKernel(period)
	rawTrend := centeredMovingAverage(points, kernel)
	trend := extrapolateMissingTrend(rawTrend)
	seasonalPattern, err := estimateAdditiveSeasonalPattern(points, rawTrend, period)
	if err != nil {
		return DecomposeResult{}, err
	}

	trendSeries := timeseriesgo.Empty()
	seasonalSeries := timeseriesgo.Empty()
	residualSeries := timeseriesgo.Empty()

	for i, point := range points {
		seasonal := seasonalPattern[i%period]
		residual := point.Value - trend[i] - seasonal

		trendSeries.AddPoint(timeseriesgo.DataPoint{
			Timestamp: point.Timestamp,
			Value:     trend[i],
		})
		seasonalSeries.AddPoint(timeseriesgo.DataPoint{
			Timestamp: point.Timestamp,
			Value:     seasonal,
		})
		residualSeries.AddPoint(timeseriesgo.DataPoint{
			Timestamp: point.Timestamp,
			Value:     residual,
		})
	}

	return DecomposeResult{
		Trend:    trendSeries,
		Seasonal: seasonalSeries,
		Residual: residualSeries,
	}, nil
}

func seasonalKernel(period int) []float64 {
	if period%2 == 1 {
		kernel := make([]float64, period)
		weight := 1.0 / float64(period)
		for i := range kernel {
			kernel[i] = weight
		}
		return kernel
	}

	kernel := make([]float64, period+1)
	scale := 1.0 / float64(period)
	kernel[0] = 0.5 * scale
	kernel[len(kernel)-1] = 0.5 * scale
	for i := 1; i < len(kernel)-1; i++ {
		kernel[i] = scale
	}
	return kernel
}

func centeredMovingAverage(points []timeseriesgo.DataPoint, kernel []float64) []float64 {
	result := make([]float64, len(points))
	for i := range result {
		result[i] = math.NaN()
	}

	half := len(kernel) / 2
	for center := half; center < len(points)-half; center++ {
		sum := 0.0
		for k, weight := range kernel {
			sum += weight * points[center-half+k].Value
		}
		result[center] = sum
	}

	return result
}

func extrapolateMissingTrend(rawTrend []float64) []float64 {
	trend := make([]float64, len(rawTrend))
	copy(trend, rawTrend)

	firstValid := -1
	secondValid := -1
	lastValid := -1
	previousValid := -1

	for i, value := range rawTrend {
		if math.IsNaN(value) {
			continue
		}
		if firstValid == -1 {
			firstValid = i
		} else if secondValid == -1 {
			secondValid = i
		}
		previousValid = lastValid
		lastValid = i
	}

	if firstValid == -1 {
		return trend
	}

	if secondValid == -1 {
		for i := range trend {
			trend[i] = rawTrend[firstValid]
		}
		return trend
	}

	leftSlope := (rawTrend[secondValid] - rawTrend[firstValid]) / float64(secondValid-firstValid)
	for i := firstValid - 1; i >= 0; i-- {
		trend[i] = rawTrend[firstValid] - leftSlope*float64(firstValid-i)
	}

	if previousValid == -1 {
		previousValid = firstValid
	}
	rightSlope := (rawTrend[lastValid] - rawTrend[previousValid]) / float64(lastValid-previousValid)
	for i := lastValid + 1; i < len(trend); i++ {
		trend[i] = rawTrend[lastValid] + rightSlope*float64(i-lastValid)
	}

	return trend
}

func estimateAdditiveSeasonalPattern(points []timeseriesgo.DataPoint, rawTrend []float64, period int) ([]float64, error) {
	pattern := make([]float64, period)
	counts := make([]int, period)

	for i, point := range points {
		if math.IsNaN(rawTrend[i]) {
			continue
		}
		position := i % period
		pattern[position] += point.Value - rawTrend[i]
		counts[position]++
	}

	for i, count := range counts {
		if count == 0 {
			return nil, errors.New("not enough data to estimate seasonal component")
		}
		pattern[i] /= float64(count)
	}

	meanSeasonal := 0.0
	for _, value := range pattern {
		meanSeasonal += value
	}
	meanSeasonal /= float64(len(pattern))

	for i := range pattern {
		pattern[i] -= meanSeasonal
	}

	return pattern, nil
}
