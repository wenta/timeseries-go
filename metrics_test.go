package timeseriesgo

import (
	"testing"
	"time"
)

func TestMSE(t *testing.T) {
	ts1 := Empty()
	ts2 := Empty()
	expectedMSE := 9.0

	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 15.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 25.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 35.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), 45.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), 55.0})

	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 18.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 22.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 38.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), 42.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), 52.0})

	mse, err := MSE(ts1, ts2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if mse != expectedMSE {
		t.Errorf("Expected MSE %f, got %f", expectedMSE, mse)
	}
}

func TestRMSE(t *testing.T) {
	ts1 := Empty()
	ts2 := Empty()
	expectedRMSE := 3.0

	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 15.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 25.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 35.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), 45.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), 55.0})

	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 18.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 22.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 38.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), 42.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), 52.0})
	rmse, err := RMSE(ts1, ts2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if rmse != expectedRMSE {
		t.Errorf("Expected RMSE %f, got %f", expectedRMSE, rmse)
	}
}

func TestMAE(t *testing.T) {
	ts1 := Empty()
	ts2 := Empty()
	expected := 1.0

	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), 1.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 2.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 3.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), 2.0})
	ts1.AddPoint(DataPoint{time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), 4.0})

	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC), 2.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), 4.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 13, 0, 0, 0, time.UTC), 5.0})
	ts2.AddPoint(DataPoint{time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC), 5.0})
	mae, err := MAE(ts1, ts2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if mae != expected {
		t.Errorf("Expected MAE %f, got %f", expected, mae)
	}
}
