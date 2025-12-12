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
