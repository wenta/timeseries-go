package timeseriesgo

import "time"

/**
 * Implements a naive forecasting method that uses the last observed value to forecast future values.
 *
 * @param ts The TimeSeries to forecast.
 * @param forecastHorizon The number of future points to forecast.
 * @return A TimeSeries containing the forecasted points. Please use ts.Merge(forecast) to combine with the original series.
 */
func Naive(ts TimeSeries, forecastHorizon int) TimeSeries {
	if ts.IsEmpty() || forecastHorizon <= 0 {
		return Empty()
	}
	lastPoint, err := ts.Last()
	if err != nil {
		return Empty()
	}
	forecastSeries := Empty()
	interval := ts.datapoints[1].timestamp.Sub(ts.datapoints[0].timestamp)
	for i := 1; i <= forecastHorizon; i++ {
		forecastTime := lastPoint.timestamp.Add(time.Duration(i) * interval)
		forecastSeries.AddPoint(DataPoint{
			timestamp: forecastTime,
			value:     lastPoint.value,
		})
	}
	return forecastSeries
}
