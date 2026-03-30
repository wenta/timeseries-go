package tsio

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

/**
 * Parses a CSV reader into a TimeSeries using a custom time format.
 * Expected columns per row: timestamp, value (float64). No header support.
 *
 * @param reader The CSV reader containing the input rows.
 * @param timeFormat The Go time layout used to parse the timestamp column.
 * @param label The label to assign to the resulting TimeSeries.
 *
 * @return A parsed TimeSeries, or an error if reading or parsing fails.
 */
func FromStringWithTimeFormat(reader csv.Reader, timeFormat string, label string) (timeseriesgo.TimeSeries, error) {
	data, err := reader.ReadAll()
	if err != nil {
		return timeseriesgo.EmptyLabeled(label), err
	}

	ts := timeseriesgo.Empty()
	for _, row := range data {
		if len(row) != 2 {
			return timeseriesgo.Empty(), errors.New("expected exactly 2 columns per row")
		}

		tsStr := row[0]
		valStr := row[1]

		dt, err := time.Parse(timeFormat, tsStr)
		if err != nil {
			return timeseriesgo.Empty(), err
		}

		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return timeseriesgo.Empty(), err
		}

		ts.AddPoint(timeseriesgo.DataPoint{Timestamp: dt, Value: val})
	}

	return ts, nil
}

/**
 * Parses a CSV reader into a TimeSeries using RFC3339 timestamps.
 * Expected columns per row: timestamp (RFC3339), value (float64). No header support.
 *
 * @param reader The CSV reader containing the input rows.
 * @param label The label to assign to the resulting TimeSeries.
 *
 * @return A parsed TimeSeries, or an error if reading or parsing fails.
 */
func FromString(reader csv.Reader, label string) (timeseriesgo.TimeSeries, error) {
	timeFormat := time.RFC3339
	return FromStringWithTimeFormat(reader, timeFormat, label)
}

/**
 * Serializes a TimeSeries to CSV string using a custom time format.
 * Output columns per row: timestamp, value (float64). No header.
 *
 * @param ts The TimeSeries to serialize.
 * @param timeFormat The Go time layout used to format the timestamp column.
 *
 * @return A CSV string representation of the TimeSeries, or an error if writing fails.
 */
func ToStringWithTimeFormat(ts timeseriesgo.TimeSeries, timeFormat string) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	for _, dp := range ts.DataPoints() {
		row := []string{
			dp.Timestamp.Format(timeFormat),
			strconv.FormatFloat(dp.Value, 'f', -1, 64),
		}
		if err := w.Write(row); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

/**
 * Serializes a TimeSeries to CSV string using RFC3339 timestamps.
 * Output columns per row: timestamp, value (float64). No header.
 *
 * @param ts The TimeSeries to serialize.
 *
 * @return A CSV string representation of the TimeSeries, or an error if writing fails.
 */
func ToString(ts timeseriesgo.TimeSeries) (string, error) {
	timeFormat := time.RFC3339
	return ToStringWithTimeFormat(ts, timeFormat)
}
