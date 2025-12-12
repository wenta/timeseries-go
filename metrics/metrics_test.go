package metrics

import (
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

func TestMSE(t *testing.T) {
	ts1 := timeseriesgo.Empty()
	ts2 := timeseriesgo.Empty()
	expectedMSE := 9.0

	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), Value: 15.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), Value: 25.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), Value: 35.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), Value: 45.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), Value: 55.0})

	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), Value: 18.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), Value: 22.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), Value: 38.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), Value: 42.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), Value: 52.0})

	mse, err := MSE(ts1, ts2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if mse != expectedMSE {
		t.Errorf("Expected MSE %f, got %f", expectedMSE, mse)
	}
}

func TestRMSE(t *testing.T) {
	ts1 := timeseriesgo.Empty()
	ts2 := timeseriesgo.Empty()
	expectedRMSE := 3.0

	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), Value: 15.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), Value: 25.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), Value: 35.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), Value: 45.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), Value: 55.0})

	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), Value: 18.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), Value: 22.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), Value: 38.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), Value: 42.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), Value: 52.0})
	rmse, err := RMSE(ts1, ts2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if rmse != expectedRMSE {
		t.Errorf("Expected RMSE %f, got %f", expectedRMSE, rmse)
	}
}

func TestMAE(t *testing.T) {
	ts1 := timeseriesgo.Empty()
	ts2 := timeseriesgo.Empty()
	expected := 1.0

	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), Value: 1.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), Value: 2.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), Value: 3.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), Value: 2.0})
	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), Value: 4.0})

	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), Value: 2.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), Value: 4.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), Value: 5.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), Value: 5.0})
	mae, err := MAE(ts1, ts2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if mae != expected {
		t.Errorf("Expected MAE %f, got %f", expected, mae)
	}
}

func TestMSE_NoOverlapReturnsError(t *testing.T) {
	ts1 := timeseriesgo.Empty()
	ts2 := timeseriesgo.Empty()

	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), Value: 1.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 2, 10, 0, 0, 0, time.UTC), Value: 1.0})

	_, err := MSE(ts1, ts2)
	if err == nil {
		t.Fatalf("Expected error for non-overlapping series, got nil")
	}
}

func TestMAE_NoOverlapReturnsError(t *testing.T) {
	ts1 := timeseriesgo.Empty()
	ts2 := timeseriesgo.Empty()

	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), Value: 1.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 2, 10, 0, 0, 0, time.UTC), Value: 1.0})

	_, err := MAE(ts1, ts2)
	if err == nil {
		t.Fatalf("Expected error for non-overlapping series, got nil")
	}
}

func TestRMSE_NoOverlapReturnsError(t *testing.T) {
	ts1 := timeseriesgo.Empty()
	ts2 := timeseriesgo.Empty()

	ts1.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), Value: 1.0})
	ts2.AddPoint(timeseriesgo.DataPoint{Timestamp: time.Date(2024, 6, 2, 10, 0, 0, 0, time.UTC), Value: 1.0})

	_, err := RMSE(ts1, ts2)
	if err == nil {
		t.Fatalf("Expected error for non-overlapping series, got nil")
	}
}
