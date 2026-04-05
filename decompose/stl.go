package decompose

import (
	"errors"
	"math"
	"sort"

	timeseriesgo "github.com/wenta/timeseries-go"
)

// STLConfig configures STL decomposition.
type STLConfig struct {
	// Period is the number of observations in one full seasonal cycle.
	// Examples include 12 for monthly yearly seasonality or 24 for hourly daily seasonality.
	Period int
	// Seasonal controls the LOESS window length used for seasonal smoothing.
	// Zero uses the default value of 7.
	Seasonal int
	// Trend controls the LOESS window length used for trend smoothing.
	// Zero derives the default from Period and Seasonal using the STL heuristic.
	Trend int
	// LowPass controls the low-pass smoother length used to separate the seasonal
	// component from slowly varying drift in the subseries smoothing step.
	// Zero derives the default from Period.
	LowPass int
	// Robust enables residual-based reweighting so large outliers influence the fit less.
	Robust bool
	// InnerIterations controls how many seasonal/trend refinement steps are performed per outer pass.
	// Zero uses the default value of 2.
	InnerIterations int
	// OuterIterations controls how many robust reweighting passes are performed when Robust is enabled.
	// Zero uses the default value of 15 when Robust is true, otherwise 0.
	OuterIterations int
}

/**
 * Decomposes a TimeSeries using STL (Season-Trend decomposition using LOESS).
 *
 * This implementation follows the STL procedure introduced by Cleveland,
 * Cleveland, McRae, and Terpenning (1990): repeated LOESS smoothing is used
 * to estimate seasonal and trend components, and an optional robust mode
 * downweights outliers using residual-based weights.
 *
 * @param ts The TimeSeries to decompose. Expected that ts is already sorted by timestamp.
 * @param config The STL configuration. Period is required, while zero-valued smoother lengths and iteration counts use defaults.
 *
 * @return A DecomposeResult containing full-length Trend, Seasonal, and Residual series.
 * @return An error if the series is empty or the configuration is invalid.
 */
func STL(ts timeseriesgo.TimeSeries, config STLConfig) (DecomposeResult, error) {
	if ts.IsEmpty() {
		return DecomposeResult{}, errors.New("TimeSeries is empty")
	}

	cfg, err := resolveSTLConfig(ts.Length(), config)
	if err != nil {
		return DecomposeResult{}, err
	}

	points := ts.DataPoints()
	values := make([]float64, len(points))
	for i, point := range points {
		values[i] = point.Value
	}

	initial, err := SeasonalDecompose(ts, cfg.Period)
	if err != nil {
		return DecomposeResult{}, err
	}
	trend := initial.Trend.Values()
	seasonal := initial.Seasonal.Values()
	residual := make([]float64, len(values))
	detrended := make([]float64, len(values))
	deseasonalized := make([]float64, len(values))
	lowPassWork1 := make([]float64, len(values))
	lowPassWork2 := make([]float64, len(values))
	lowPassWork3 := make([]float64, len(values))
	seasonalFilter := seasonalKernel(cfg.Period)
	threePointFilter := []float64{1.0 / 3, 1.0 / 3, 1.0 / 3}
	var robustWeights []float64
	if cfg.Robust {
		robustWeights = ones(len(values))
	}

	for outer := 0; outer <= cfg.OuterIterations; outer++ {
		for inner := 0; inner < cfg.InnerIterations; inner++ {
			subtractInto(detrended, values, trend)
			subseriesSmooth := smoothSeasonalSubseries(detrended, robustWeights, cfg.Period, cfg.Seasonal)
			lowPass := stlLowPass(subseriesSmooth, cfg.LowPass, robustWeights, seasonalFilter, threePointFilter, lowPassWork1, lowPassWork2, lowPassWork3)
			subtractInto(seasonal, subseriesSmooth, lowPass)
			subtractInto(deseasonalized, values, seasonal)
			trend = loessSmooth(deseasonalized, cfg.Trend, robustWeights)
		}

		subtract3Into(residual, values, trend, seasonal)
		if !cfg.Robust || outer == cfg.OuterIterations {
			break
		}
		robustWeights = bisquareWeights(residual)
	}

	return DecomposeResult{
		Trend:    valuesToSeries(points, trend),
		Seasonal: valuesToSeries(points, seasonal),
		Residual: valuesToSeries(points, residual),
	}, nil
}

func resolveSTLConfig(seriesLength int, input STLConfig) (STLConfig, error) {
	if input.Period < 2 {
		return STLConfig{}, errors.New("period must be at least 2")
	}
	if seriesLength < 2*input.Period {
		return STLConfig{}, errors.New("STL requires at least two complete cycles")
	}

	cfg := input

	if cfg.Seasonal == 0 {
		cfg.Seasonal = maxInt(7, smallestOddAtLeast(7))
	}
	if cfg.Trend == 0 {
		estimate := 1.5 * float64(cfg.Period) / (1.0 - 1.5/float64(cfg.Seasonal))
		cfg.Trend = smallestOddGreaterThan(estimate)
	}
	if cfg.LowPass == 0 {
		cfg.LowPass = smallestOddGreaterThan(float64(cfg.Period))
	}
	if cfg.InnerIterations == 0 {
		cfg.InnerIterations = 2
	}
	if cfg.Robust {
		if cfg.OuterIterations == 0 {
			cfg.OuterIterations = 15
		}
	}

	if cfg.Seasonal < 3 || cfg.Trend < 3 || cfg.LowPass < 3 {
		return STLConfig{}, errors.New("seasonal, trend, and low-pass lengths must be at least 3")
	}
	if cfg.Seasonal%2 == 0 || cfg.Trend%2 == 0 || cfg.LowPass%2 == 0 {
		return STLConfig{}, errors.New("seasonal, trend, and low-pass lengths must be odd")
	}
	if cfg.InnerIterations < 1 {
		return STLConfig{}, errors.New("inner iterations must be at least 1")
	}
	if cfg.OuterIterations < 0 {
		return STLConfig{}, errors.New("outer iterations cannot be negative")
	}

	return cfg, nil
}

func smoothSeasonalSubseries(values []float64, robustWeights []float64, period int, seasonalWindow int) []float64 {
	seasonal := make([]float64, len(values))

	for offset := 0; offset < period; offset++ {
		subseriesLength := (len(values) - offset + period - 1) / period
		if subseriesLength < 0 {
			subseriesLength = 0
		}
		subseriesValues := make([]float64, 0, subseriesLength)
		indices := make([]int, 0, subseriesLength)
		var subseriesWeights []float64
		if robustWeights != nil {
			subseriesWeights = make([]float64, 0, subseriesLength)
		}

		for idx := offset; idx < len(values); idx += period {
			subseriesValues = append(subseriesValues, values[idx])
			if robustWeights != nil {
				subseriesWeights = append(subseriesWeights, robustWeights[idx])
			}
			indices = append(indices, idx)
		}

		smoothed := loessSmooth(subseriesValues, seasonalWindow, subseriesWeights)
		for i, idx := range indices {
			seasonal[idx] = smoothed[i]
		}
	}

	return seasonal
}

func stlLowPass(values []float64, lowPassWindow int, robustWeights []float64, seasonalFilter []float64, threePointFilter []float64, work1 []float64, work2 []float64, work3 []float64) []float64 {
	centeredMovingAverageInto(work1, values, seasonalFilter)
	extrapolateMissingTrendInPlace(work1)
	centeredMovingAverageInto(work2, work1, seasonalFilter)
	extrapolateMissingTrendInPlace(work2)
	centeredMovingAverageInto(work3, work2, threePointFilter)
	extrapolateMissingTrendInPlace(work3)
	return loessSmooth(work3, lowPassWindow, robustWeights)
}

func centeredMovingAverageValues(values []float64, kernel []float64) []float64 {
	result := make([]float64, len(values))
	centeredMovingAverageInto(result, values, kernel)
	return result
}

func centeredMovingAverageInto(dst []float64, values []float64, kernel []float64) {
	for i := range dst {
		dst[i] = math.NaN()
	}

	half := len(kernel) / 2
	for center := half; center < len(values)-half; center++ {
		sum := 0.0
		for k, weight := range kernel {
			sum += weight * values[center-half+k]
		}
		dst[center] = sum
	}
}

func loessSmooth(values []float64, window int, robustWeights []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	if len(values) == 1 {
		return []float64{values[0]}
	}

	window = minInt(window, len(values))
	if window < 2 {
		window = 2
	}

	result := make([]float64, len(values))
	for i := range values {
		left, right := loessWindowBounds(len(values), window, i)
		maxDistance := math.Max(float64(i-left), float64(right-i))
		if maxDistance == 0 {
			result[i] = values[i]
			continue
		}

		sumW := 0.0
		sumWX := 0.0
		sumWY := 0.0
		sumWXX := 0.0
		sumWXY := 0.0

		for j := left; j <= right; j++ {
			distance := math.Abs(float64(j-i)) / maxDistance
			weight := tricube(distance)
			if robustWeights != nil {
				weight *= robustWeights[j]
			}
			if weight == 0 {
				continue
			}

			x := float64(j - i)
			y := values[j]
			sumW += weight
			sumWX += weight * x
			sumWY += weight * y
			sumWXX += weight * x * x
			sumWXY += weight * x * y
		}

		if sumW == 0 {
			result[i] = values[i]
			continue
		}

		denominator := sumW*sumWXX - sumWX*sumWX
		if math.Abs(denominator) <= 1e-12 {
			result[i] = sumWY / sumW
			continue
		}

		slope := (sumW*sumWXY - sumWX*sumWY) / denominator
		intercept := (sumWY - slope*sumWX) / sumW
		result[i] = intercept
	}

	return result
}

func loessWindowBounds(n int, window int, center int) (int, int) {
	left := center - window/2
	right := left + window - 1

	if left < 0 {
		right -= left
		left = 0
	}
	if right >= n {
		left -= right - n + 1
		right = n - 1
		if left < 0 {
			left = 0
		}
	}

	return left, right
}

func tricube(x float64) float64 {
	if x >= 1 {
		return 0
	}
	value := 1 - x*x*x
	return value * value * value
}

func bisquareWeights(residual []float64) []float64 {
	weights := make([]float64, len(residual))
	absResidual := make([]float64, len(residual))
	for i, value := range residual {
		absResidual[i] = math.Abs(value)
	}

	scale := 6 * median(absResidual)
	if scale <= 1e-12 {
		for i := range weights {
			weights[i] = 1
		}
		return weights
	}

	for i, value := range residual {
		r := math.Abs(value) / scale
		if r >= 1 {
			weights[i] = 0
			continue
		}
		w := 1 - r*r
		weights[i] = w * w
	}

	return weights
}

func median(values []float64) float64 {
	cp := make([]float64, len(values))
	copy(cp, values)
	sort.Float64s(cp)

	mid := len(cp) / 2
	if len(cp)%2 == 1 {
		return cp[mid]
	}
	return (cp[mid-1] + cp[mid]) / 2
}

func valuesToSeries(points []timeseriesgo.DataPoint, values []float64) timeseriesgo.TimeSeries {
	result := make([]timeseriesgo.DataPoint, len(points))
	for i, point := range points {
		result[i] = timeseriesgo.DataPoint{
			Timestamp: point.Timestamp,
			Value:     values[i],
		}
	}
	return timeseriesgo.FromDataPoints(result)
}

func subtract(left []float64, right []float64) []float64 {
	result := make([]float64, len(left))
	subtractInto(result, left, right)
	return result
}

func subtractInto(dst []float64, left []float64, right []float64) {
	for i := range left {
		dst[i] = left[i] - right[i]
	}
}

func subtract3Into(dst []float64, source []float64, subtractA []float64, subtractB []float64) {
	for i := range source {
		dst[i] = source[i] - subtractA[i] - subtractB[i]
	}
}

func extrapolateMissingTrendInPlace(trend []float64) {
	firstValid := -1
	secondValid := -1
	lastValid := -1
	previousValid := -1

	for i, value := range trend {
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
		return
	}

	if secondValid == -1 {
		for i := range trend {
			trend[i] = trend[firstValid]
		}
		return
	}

	leftSlope := (trend[secondValid] - trend[firstValid]) / float64(secondValid-firstValid)
	for i := firstValid - 1; i >= 0; i-- {
		trend[i] = trend[firstValid] - leftSlope*float64(firstValid-i)
	}

	if previousValid == -1 {
		previousValid = firstValid
	}
	rightSlope := (trend[lastValid] - trend[previousValid]) / float64(lastValid-previousValid)
	for i := lastValid + 1; i < len(trend); i++ {
		trend[i] = trend[lastValid] + rightSlope*float64(i-lastValid)
	}
}

func ones(n int) []float64 {
	result := make([]float64, n)
	for i := range result {
		result[i] = 1
	}
	return result
}

func smallestOddGreaterThan(value float64) int {
	candidate := int(math.Floor(value)) + 1
	if candidate%2 == 0 {
		candidate++
	}
	return candidate
}

func smallestOddAtLeast(value int) int {
	if value%2 == 1 {
		return value
	}
	return value + 1
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
