package forecast

import (
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
