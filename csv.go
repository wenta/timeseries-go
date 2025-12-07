package timeseriesgo

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"
	"time"
)

/**
 * Parses a CSV reader into a TimeSeries.
 * Expected columns per row: timestamp, value (float64). No header support.
 */
func FromStringWithTimeFormat(reader csv.Reader, timeFormat string) (TimeSeries, error) {
	data, err := reader.ReadAll()
	if err != nil {
		return Empty(), err
	}

	ts := Empty()
	for _, row := range data {
		if len(row) != 2 {
			return Empty(), errors.New("expected exactly 2 columns per row")
		}

		tsStr := row[0]
		valStr := row[1]

		dt, err := time.Parse(timeFormat, tsStr)
		if err != nil {
			return Empty(), err
		}

		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return Empty(), err
		}

		ts.AddPoint(DataPoint{timestamp: dt, value: val})
	}

	return ts, nil
}

/**
 * Parses a CSV reader into a TimeSeries.
 * Expected columns per row: timestamp (RFC3339), value (float64). No header support.
 */
func FromString(reader csv.Reader) (TimeSeries, error) {
	timeFormat := time.RFC3339
	return FromStringWithTimeFormat(reader, timeFormat)
}

func ToStringWithTimeFormat(ts TimeSeries, timeFormat string) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	for _, dp := range ts.datapoints {
		row := []string{
			dp.timestamp.Format(timeFormat),
			strconv.FormatFloat(dp.value, 'f', -1, 64),
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
func ToString(ts TimeSeries) (string, error) {
	timeFormat := time.RFC3339
	return ToStringWithTimeFormat(ts, timeFormat)
}
