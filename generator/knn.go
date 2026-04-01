package generator

import (
	"math"
	"math/rand/v2"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

/**
 * Resamples a univariate TimeSeries by matching recent windows against similar historical windows.
 *
 * The generator starts from the last window values of the source series, finds the
 * k nearest historical windows using Euclidean distance, and samples the next value
 * from one of those neighbors.
 *
 * @param ts The source TimeSeries providing historical windows and next-step values.
 * @param outputIndex The timestamps to use for the generated output series.
 * @param window The number of trailing values used as the matching context.
 * @param k The number of nearest neighbors eligible for sampling.
 *
 * @return A TimeSeries with timestamps from outputIndex and values generated from historical nearest-neighbor continuations.
 */
func ResampleKNN(ts timeseriesgo.TimeSeries, outputIndex []time.Time, window int, k int) timeseriesgo.TimeSeries {
	return ResampleKNNWithRand(ts, outputIndex, window, k, nil)
}

/**
 * Resamples a univariate TimeSeries by matching recent windows against similar historical windows using a caller-provided RNG.
 *
 * @param ts The source TimeSeries providing historical windows and next-step values.
 * @param outputIndex The timestamps to use for the generated output series.
 * @param window The number of trailing values used as the matching context.
 * @param k The number of nearest neighbors eligible for sampling.
 * @param rng The random source used to choose among the nearest neighbors. Nil uses the package default RNG.
 *
 * @return A TimeSeries with timestamps from outputIndex and values generated from historical nearest-neighbor continuations.
 */
func ResampleKNNWithRand(ts timeseriesgo.TimeSeries, outputIndex []time.Time, window int, k int, rng *rand.Rand) timeseriesgo.TimeSeries {
	if ts.IsEmpty() || len(outputIndex) == 0 || window <= 0 || k <= 0 {
		return timeseriesgo.Empty()
	}

	sourceValues := ts.Values()
	if len(sourceValues) <= window {
		return timeseriesgo.Empty()
	}

	context := append([]float64(nil), sourceValues[len(sourceValues)-window:]...)
	values := make([]float64, len(outputIndex))

	candidateCount := len(sourceValues) - window
	limit := minInt(k, candidateCount)
	for i := range outputIndex {
		neighbors := make([]knnNeighbor, 0, candidateCount)
		for start := 0; start < candidateCount; start++ {
			distance := euclideanDistance(context, sourceValues[start:start+window])
			neighbors = append(neighbors, knnNeighbor{
				distance: distance,
				next:     sourceValues[start+window],
			})
		}

		sortNeighbors(neighbors)
		chosen := neighbors[randomIntN(rng, limit)]
		values[i] = chosen.next
		context = append(context[1:], chosen.next)
	}

	return zipIndexValues(outputIndex, values)
}

type knnNeighbor struct {
	distance float64
	next     float64
}

func euclideanDistance(left []float64, right []float64) float64 {
	sum := 0.0
	for i := range left {
		diff := left[i] - right[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

func sortNeighbors(neighbors []knnNeighbor) {
	for i := 1; i < len(neighbors); i++ {
		j := i
		for j > 0 && neighbors[j].distance < neighbors[j-1].distance {
			neighbors[j], neighbors[j-1] = neighbors[j-1], neighbors[j]
			j--
		}
	}
}
