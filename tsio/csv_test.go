package tsio

import (
	"encoding/csv"
	"strings"
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

func TestFromStringParsesRows(t *testing.T) {
	input := "2024-06-01T00:00:00Z,1.5\n2024-06-01T00:01:00Z,2.5\n"
	reader := csv.NewReader(strings.NewReader(input))

	ts, err := FromString(*reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.Length() != 2 {
		t.Fatalf("expected 2 rows, got %d", ts.Length())
	}

	expectedTimes := []time.Time{
		time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 6, 1, 0, 1, 0, 0, time.UTC),
	}
	expectedValues := []float64{1.5, 2.5}

	for i, dp := range ts.DataPoints() {
		if !dp.Timestamp.Equal(expectedTimes[i]) {
			t.Errorf("row %d timestamp mismatch: expected %v, got %v", i, expectedTimes[i], dp.Timestamp)
		}
		if dp.Value != expectedValues[i] {
			t.Errorf("row %d value mismatch: expected %v, got %v", i, expectedValues[i], dp.Value)
		}
	}
}

func TestToStringProducesCSV(t *testing.T) {
	ts := timeseriesgo.Empty()
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), Value: 1.5})
	ts.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 0, 1, 0, 0, time.UTC), Value: 2.5})

	out, err := ToString(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "2024-06-01T00:00:00Z,1.5\n2024-06-01T00:01:00Z,2.5\n"
	if out != expected {
		t.Errorf("expected CSV output:\n%q\ngot:\n%q", expected, out)
	}
}
