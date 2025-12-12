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
 * Parses a CSV reader into a TimeSeries.
 * Expected columns per row: timestamp, value (float64). No header support.
 */
func FromStringWithTimeFormat(reader csv.Reader, timeFormat string) (timeseriesgo.TimeSeries, error) {
	data, err := reader.ReadAll()
	if err != nil {
		return timeseriesgo.Empty(), err
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
 * Parses a CSV reader into a TimeSeries.
 * Expected columns per row: timestamp (RFC3339), value (float64). No header support.
 */
func FromString(reader csv.Reader) (timeseriesgo.TimeSeries, error) {
	timeFormat := time.RFC3339
	return FromStringWithTimeFormat(reader, timeFormat)
}

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
 * Serializes a TimeSeries to CSV string (timestamp RFC3339, value float64). No header.
 */
func ToString(ts timeseriesgo.TimeSeries) (string, error) {
	timeFormat := time.RFC3339
	return ToStringWithTimeFormat(ts, timeFormat)
}
