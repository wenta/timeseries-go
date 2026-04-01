package generator

import (
	"math"
	"math/rand/v2"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

// Pulse describes a fixed-amplitude event rendered over a contiguous index range.
type Pulse struct {
	// StartIndex is the index position where the pulse begins.
	StartIndex int
	// Duration is the number of consecutive points affected by the pulse.
	Duration int
	// Amplitude is the value added while the pulse is active.
	Amplitude float64
}

/**
 * Generates a zero-baseline TimeSeries with additive pulses placed on the provided index.
 *
 * @param index The timestamps used for the generated series.
 * @param pulses The pulses to render on top of the zero baseline.
 *
 * @return A TimeSeries containing the rendered pulse train. Overlapping pulses add together.
 */
func PulseTrain(index []time.Time, pulses []Pulse) timeseriesgo.TimeSeries {
	if len(index) == 0 {
		return timeseriesgo.Empty()
	}

	values := make([]float64, len(index))
	for _, pulse := range pulses {
		if pulse.Duration <= 0 {
			continue
		}

		start := maxInt(pulse.StartIndex, 0)
		end := minInt(pulse.StartIndex+pulse.Duration, len(index))
		for i := start; i < end; i++ {
			values[i] += pulse.Amplitude
		}
	}

	return zipIndexValues(index, values)
}

/**
 * Generates ascending event indices using a discrete Poisson-style process.
 *
 * The implementation samples each index independently with probability 1-exp(-lambda),
 * which corresponds to the chance of at least one event in a unit interval for a
 * Poisson process with rate lambda.
 *
 * @param length The exclusive upper bound of valid returned indices.
 * @param lambda The expected event rate per index position.
 *
 * @return A sorted slice of valid event indices in the range [0, length).
 */
func PoissonEventIndices(length int, lambda float64) []int {
	if length <= 0 || lambda <= 0 {
		return nil
	}

	probability := 1 - math.Exp(-lambda)
	indices := make([]int, 0)
	for i := 0; i < length; i++ {
		if rand.Float64() < probability {
			indices = append(indices, i)
		}
	}
	return indices
}
