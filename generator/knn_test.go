package generator

import (
	"math/rand/v2"
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

func TestResampleKNNDeterministicRepeatingPattern(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	sourceIndex := MakeSeriesIndex(base, time.Minute, 6)
	sourceValues := []float64{1, 2, 1, 2, 1, 2}
	source, err := timeseriesgo.Zip(sourceIndex, sourceValues)
	if err != nil {
		t.Fatalf("unexpected error creating source: %v", err)
	}

	outputIndex := MakeSeriesIndex(base.Add(time.Hour), time.Minute, 4)
	result := ResampleKNN(source, outputIndex, 2, 1)

	expectedValues := []float64{1, 2, 1, 2}
	if result.Length() != len(outputIndex) {
		t.Fatalf("expected output length %d, got %d", len(outputIndex), result.Length())
	}

	for i, dp := range result.DataPoints() {
		if !dp.Timestamp.Equal(outputIndex[i]) {
			t.Fatalf("timestamp at index %d: expected %v, got %v", i, outputIndex[i], dp.Timestamp)
		}
		if dp.Value != expectedValues[i] {
			t.Fatalf("value at index %d: expected %f, got %f", i, expectedValues[i], dp.Value)
		}
	}
}

func TestResampleKNNInvalidInput(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := MakeSeriesIndex(base, time.Minute, 3)
	source, err := timeseriesgo.Zip(index, []float64{1, 2, 3})
	if err != nil {
		t.Fatalf("unexpected error creating source: %v", err)
	}

	cases := []timeseriesgo.TimeSeries{
		ResampleKNN(timeseriesgo.Empty(), index, 2, 1),
		ResampleKNN(source, nil, 2, 1),
		ResampleKNN(source, index, 0, 1),
		ResampleKNN(source, index, 2, 0),
		ResampleKNN(source, index, 3, 1),
	}

	for i, result := range cases {
		if !result.IsEmpty() {
			t.Fatalf("case %d: expected empty result, got length %d", i, result.Length())
		}
	}
}

func TestResampleKNNWithRandIsReproducible(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	sourceIndex := MakeSeriesIndex(base, time.Minute, 8)
	sourceValues := []float64{1, 2, 3, 2, 1, 2, 3, 2}
	source, err := timeseriesgo.Zip(sourceIndex, sourceValues)
	if err != nil {
		t.Fatalf("unexpected error creating source: %v", err)
	}

	outputIndex := MakeSeriesIndex(base.Add(time.Hour), time.Minute, 6)
	rngA := rand.New(rand.NewPCG(7, 11))
	rngB := rand.New(rand.NewPCG(7, 11))

	left := ResampleKNNWithRand(source, outputIndex, 2, 2, rngA)
	right := ResampleKNNWithRand(source, outputIndex, 2, 2, rngB)

	leftPoints := left.DataPoints()
	rightPoints := right.DataPoints()
	if len(leftPoints) != len(rightPoints) {
		t.Fatalf("expected equal lengths, got %d and %d", len(leftPoints), len(rightPoints))
	}
	for i := range leftPoints {
		if leftPoints[i] != rightPoints[i] {
			t.Fatalf("expected reproducible outputs, mismatch at %d: %+v vs %+v", i, leftPoints[i], rightPoints[i])
		}
	}
}
