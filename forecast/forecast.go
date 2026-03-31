package forecast

import (
	"math"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

/**
 * Implements a naive forecasting method that uses the last observed value to forecast future values.
 *
 * @param ts The TimeSeries to forecast.
 * @param forecastHorizon The number of future points to forecast.
 * @return A TimeSeries containing the forecasted points. Please use ts.Merge(forecast) to combine with the original series.
 */
func Naive(ts timeseriesgo.TimeSeries, forecastHorizon int) timeseriesgo.TimeSeries {
	if ts.IsEmpty() || forecastHorizon <= 0 {
		return timeseriesgo.Empty()
	}
	lastPoint, err := ts.Last()
	if err != nil {
		return timeseriesgo.Empty()
	}
	forecastSeries := timeseriesgo.Empty()
	points := ts.DataPoints()
	if len(points) < 2 {
		return forecastSeries
	}
	interval := points[1].Timestamp.Sub(points[0].Timestamp)
	for i := 1; i <= forecastHorizon; i++ {
		forecastTime := lastPoint.Timestamp.Add(time.Duration(i) * interval)
		forecastSeries.AddPoint(timeseriesgo.DataPoint{
			Timestamp: forecastTime,
			Value:     lastPoint.Value,
		})
	}
	return forecastSeries
}

/**
 * Implements Simple Exponential Smoothing (SES) forecasting method.
 *
 * @param ts The TimeSeries to forecast. Expected that ts is already sorted by timestamp
 * @param alpha The smoothing factor (0 < alpha <= 1).
 * @param forecastHorizon The number of future points to forecast.
 * @return A TimeSeries containing the forecasted points. Please use ts.Merge(forecast) to combine with the original series.
 */
func SimpleExponentialSmoothing(ts timeseriesgo.TimeSeries, alpha float64, forecastHorizon int) timeseriesgo.TimeSeries {
	if ts.IsEmpty() || forecastHorizon <= 0 || alpha < 0 || alpha > 1 {
		return timeseriesgo.Empty()
	}
	points := ts.DataPoints()
	if len(points) < 2 {
		return timeseriesgo.Empty()
	}

	// Initialize the smoothed value with the first data point's value.
	smoothedValue := points[0].Value

	// Apply Simple Exponential Smoothing.
	for _, point := range points {
		smoothedValue = alpha*point.Value + (1-alpha)*smoothedValue
	}

	// Generate forecasted points.
	forecastSeries := timeseriesgo.Empty()
	lastPoint, _ := ts.Last()
	interval := points[1].Timestamp.Sub(points[0].Timestamp)
	for i := 1; i <= forecastHorizon; i++ {
		forecastTime := lastPoint.Timestamp.Add(time.Duration(i) * interval)
		forecastSeries.AddPoint(timeseriesgo.DataPoint{
			Timestamp: forecastTime,
			Value:     smoothedValue,
		})
	}
	return forecastSeries
}

/**
 * Implements Holt's linear trend method (double exponential smoothing).
 *
 * @param ts The TimeSeries to forecast. Expected that ts is already sorted by timestamp
 * @param alpha The level smoothing factor (0 <= alpha <= 1).
 * @param beta The trend smoothing factor (0 <= beta <= 1).
 * @param forecastHorizon The number of future points to forecast.
 * @return A TimeSeries containing the forecasted points. Please use ts.Merge(forecast) to combine with the original series.
 */
func DoubleExponentialSmoothing(ts timeseriesgo.TimeSeries, alpha float64, beta float64, forecastHorizon int) timeseriesgo.TimeSeries {
	if ts.IsEmpty() || forecastHorizon <= 0 || alpha < 0 || alpha > 1 || beta < 0 || beta > 1 {
		return timeseriesgo.Empty()
	}

	points := ts.DataPoints()
	if len(points) < 2 {
		return timeseriesgo.Empty()
	}

	level := points[0].Value
	trend := points[1].Value - points[0].Value
	return doubleExponentialSmoothingWithInitialization(points, alpha, beta, forecastHorizon, level, trend, 1)
}

/**
 * Implements Holt's linear trend method (double exponential smoothing)
 * using estimated initial level and trend.
 *
 * For short non-seasonal series (< 10 observations), this falls back to the
 * simple initialization used by DoubleExponentialSmoothing. For longer series,
 * initial level and trend are estimated using a least-squares line fit over
 * the first 10 observations, matching the non-seasonal heuristic used by
 * statsmodels for initialization_method="estimated".
 *
 * @param ts The TimeSeries to forecast. Expected that ts is already sorted by timestamp
 * @param alpha The level smoothing factor (0 <= alpha <= 1).
 * @param beta The trend smoothing factor (0 <= beta <= 1).
 * @param forecastHorizon The number of future points to forecast.
 * @return A TimeSeries containing the forecasted points. Please use ts.Merge(forecast) to combine with the original series.
 */
func DoubleExponentialSmoothingEstimated(ts timeseriesgo.TimeSeries, alpha float64, beta float64, forecastHorizon int) timeseriesgo.TimeSeries {
	if ts.IsEmpty() || forecastHorizon <= 0 || alpha < 0 || alpha > 1 || beta < 0 || beta > 1 {
		return timeseriesgo.Empty()
	}

	points := ts.DataPoints()
	if len(points) < 2 {
		return timeseriesgo.Empty()
	}

	level, trend := estimateHoltInitialLevelTrend(points)
	return doubleExponentialSmoothingWithInitialization(points, alpha, beta, forecastHorizon, level, trend, 0)
}

/**
 * Implements Holt-Winters additive seasonal forecasting (triple exponential smoothing).
 *
 * @param ts The TimeSeries to forecast. Expected that ts is already sorted by timestamp.
 * @param alpha The level smoothing factor (0 <= alpha <= 1).
 * @param beta The trend smoothing factor (0 <= beta <= 1).
 * @param gamma The seasonal smoothing factor (0 <= gamma <= 1).
 * @param seasonLength The number of points in a full seasonal cycle.
 * @param forecastHorizon The number of future points to forecast.
 * @return A TimeSeries containing the forecasted points. Please use ts.Merge(forecast) to combine with the original series.
 */
func TripleExponentialSmoothing(ts timeseriesgo.TimeSeries, alpha float64, beta float64, gamma float64, seasonLength int, forecastHorizon int) timeseriesgo.TimeSeries {
	if ts.IsEmpty() || forecastHorizon <= 0 || seasonLength <= 0 || alpha < 0 || alpha > 1 || beta < 0 || beta > 1 || gamma < 0 || gamma > 1 {
		return timeseriesgo.Empty()
	}

	points := ts.DataPoints()
	if len(points) < 2*seasonLength || len(points) < 2 {
		return timeseriesgo.Empty()
	}

	level, trend, seasonals := estimateAdditiveSeasonalInitialization(points, seasonLength)

	for i := seasonLength; i < len(points); i++ {
		seasonal := seasonals[i%seasonLength]
		prevLevel := level

		level = alpha*(points[i].Value-seasonal) + (1-alpha)*(level+trend)
		trend = beta*(level-prevLevel) + (1-beta)*trend
		seasonals[i%seasonLength] = gamma*(points[i].Value-level) + (1-gamma)*seasonal
	}

	forecastSeries := timeseriesgo.Empty()
	lastPoint := points[len(points)-1]
	interval := points[1].Timestamp.Sub(points[0].Timestamp)
	for i := 1; i <= forecastHorizon; i++ {
		seasonal := seasonals[(len(points)+i-1)%seasonLength]
		forecastTime := lastPoint.Timestamp.Add(time.Duration(i) * interval)
		forecastSeries.AddPoint(timeseriesgo.DataPoint{
			Timestamp: forecastTime,
			Value:     level + float64(i)*trend + seasonal,
		})
	}

	return forecastSeries
}

func doubleExponentialSmoothingWithInitialization(points []timeseriesgo.DataPoint, alpha float64, beta float64, forecastHorizon int, level float64, trend float64, startIndex int) timeseriesgo.TimeSeries {
	for _, point := range points[startIndex:] {
		prevLevel := level
		level = alpha*point.Value + (1-alpha)*(level+trend)
		trend = beta*(level-prevLevel) + (1-beta)*trend
	}

	forecastSeries := timeseriesgo.Empty()
	lastPoint := points[len(points)-1]
	interval := points[1].Timestamp.Sub(points[0].Timestamp)
	for i := 1; i <= forecastHorizon; i++ {
		forecastTime := lastPoint.Timestamp.Add(time.Duration(i) * interval)
		forecastSeries.AddPoint(timeseriesgo.DataPoint{
			Timestamp: forecastTime,
			Value:     level + float64(i)*trend,
		})
	}

	return forecastSeries
}

func estimateHoltInitialLevelTrend(points []timeseriesgo.DataPoint) (float64, float64) {
	if len(points) < 10 {
		return points[0].Value, points[1].Value - points[0].Value
	}

	n := 10.0
	var sumX float64
	var sumY float64
	var sumXY float64
	var sumXX float64

	for i := range 10 {
		x := float64(i + 1)
		y := points[i].Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	denominator := n*sumXX - sumX*sumX
	if math.Abs(denominator) <= 1e-12 {
		return points[0].Value, points[1].Value - points[0].Value
	}

	trend := (n*sumXY - sumX*sumY) / denominator
	level := (sumY - trend*sumX) / n
	return level, trend
}

func estimateAdditiveSeasonalInitialization(points []timeseriesgo.DataPoint, seasonLength int) (float64, float64, []float64) {
	firstSeasonAverage := mean(points[:seasonLength])
	secondSeasonAverage := mean(points[seasonLength : 2*seasonLength])

	seasonals := make([]float64, seasonLength)
	for i := 0; i < seasonLength; i++ {
		firstSeasonDeviation := points[i].Value - firstSeasonAverage
		secondSeasonDeviation := points[i+seasonLength].Value - secondSeasonAverage
		seasonals[i] = (firstSeasonDeviation + secondSeasonDeviation) / 2
	}

	level := firstSeasonAverage
	trend := (secondSeasonAverage - firstSeasonAverage) / float64(seasonLength)
	return level, trend, seasonals
}

func mean(points []timeseriesgo.DataPoint) float64 {
	sum := 0.0
	for _, point := range points {
		sum += point.Value
	}
	return sum / float64(len(points))
}
