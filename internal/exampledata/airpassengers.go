package exampledata

import (
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

/**
 * AirPassengers returns the classic monthly airline passenger dataset from the
 * base R datasets package. Values are in thousands of passengers.
 *
 * @return A TimeSeries containing the monthly AirPassengers dataset.
 * @return An error when the embedded values cannot be zipped into a series.
 */
func AirPassengers() (timeseriesgo.TimeSeries, error) {
	values := []float64{
		112, 118, 132, 129, 121, 135, 148, 148, 136, 119, 104, 118,
		115, 126, 141, 135, 125, 149, 170, 170, 158, 133, 114, 140,
		145, 150, 178, 163, 172, 178, 199, 199, 184, 162, 146, 166,
		171, 180, 193, 181, 183, 218, 230, 242, 209, 191, 172, 194,
		196, 196, 236, 235, 229, 243, 264, 272, 237, 211, 180, 201,
		204, 188, 235, 227, 234, 264, 302, 293, 259, 229, 203, 229,
		242, 233, 267, 269, 270, 315, 364, 347, 312, 274, 237, 278,
		284, 277, 317, 313, 318, 374, 413, 405, 355, 306, 271, 306,
		315, 301, 356, 348, 355, 422, 465, 467, 404, 347, 305, 336,
		340, 318, 362, 348, 363, 435, 491, 505, 404, 359, 310, 337,
		360, 342, 406, 396, 420, 472, 548, 559, 463, 407, 362, 405,
		417, 391, 419, 461, 472, 535, 622, 606, 508, 461, 390, 432,
	}

	start := time.Date(1949, 1, 1, 0, 0, 0, 0, time.UTC)
	index := make([]time.Time, len(values))
	for i := range values {
		index[i] = start.AddDate(0, i, 0)
	}

	return timeseriesgo.Zip(index, values)
}
