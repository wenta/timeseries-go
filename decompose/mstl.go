package decompose

import (
	"errors"
	"sort"

	timeseriesgo "github.com/wenta/timeseries-go"
)

// MSTLConfig configures MSTL decomposition for multiple seasonal periods.
type MSTLConfig struct {
	// Periods lists the seasonal periods to estimate.
	// Examples include []int{24, 24*7} for hourly data with daily and weekly seasonality.
	Periods []int
	// SeasonalWindows optionally overrides the STL seasonal smoother length for each period.
	// The slice must match Periods length; zero values fall back to the default window of 7.
	SeasonalWindows []int
	// Robust enables residual-based reweighting so large outliers have less influence on the fit.
	Robust bool
	// Iterations controls how many passes MSTL performs over all seasonal periods.
	// Zero uses a default of 1 for a single period and 2 for multiple periods.
	Iterations int
	// InnerIterations is forwarded to each per-period STL fit and controls how many
	// seasonal/trend refinement steps are performed inside a single STL decomposition.
	InnerIterations int
	// OuterIterations is forwarded to each per-period STL fit and controls how many
	// robust reweighting passes are performed when Robust is enabled.
	OuterIterations int
}

// SeasonalComponent stores one seasonal component together with its period.
type SeasonalComponent struct {
	// Period is the seasonal cycle length represented by this component.
	Period int
	// Series is the extracted seasonal component for the corresponding Period.
	Series timeseriesgo.TimeSeries
}

// MSTLResult contains the components produced by MSTL.
type MSTLResult struct {
	// Trend is the long-term smoothed component after removing all estimated seasonal effects.
	Trend timeseriesgo.TimeSeries
	// Seasonal contains one extracted seasonal component for each configured period, sorted by Period.
	Seasonal []SeasonalComponent
	// Residual is the remaining component after removing trend and all seasonal components.
	Residual timeseriesgo.TimeSeries
}

/**
 * Decomposes a TimeSeries using MSTL (Multiple Seasonal-Trend decomposition using LOESS).
 *
 * MSTL extends STL to multiple seasonal periods by iteratively estimating one
 * seasonal component at a time while holding the others fixed, then smoothing
 * the remaining deseasonalized series to recover the trend.
 *
 * @param ts The TimeSeries to decompose. Expected that ts is already sorted by timestamp.
 * @param config The MSTL configuration. Periods is required; zero-valued iteration counts and seasonal windows use defaults.
 *
 * @return An MSTLResult containing the trend, one seasonal component per configured period, and the residual series.
 * @return An error if the series is empty or the configuration is invalid.
 */
func MSTL(ts timeseriesgo.TimeSeries, config MSTLConfig) (MSTLResult, error) {
	if ts.IsEmpty() {
		return MSTLResult{}, errors.New("TimeSeries is empty")
	}

	cfg, err := resolveMSTLConfig(ts.Length(), config)
	if err != nil {
		return MSTLResult{}, err
	}

	points := ts.DataPoints()
	values := make([]float64, len(points))
	for i, point := range points {
		values[i] = point.Value
	}

	seasonalValues := make([][]float64, len(cfg.Periods))
	for i := range seasonalValues {
		seasonalValues[i] = make([]float64, len(values))
	}

	iterations := cfg.Iterations
	if iterations == 0 {
		if len(cfg.Periods) == 1 {
			iterations = 1
		} else {
			iterations = 2
		}
	}

	for iteration := 0; iteration < iterations; iteration++ {
		for i, period := range cfg.Periods {
			deseasonalized := make([]float64, len(values))
			copy(deseasonalized, values)
			for j := range seasonalValues {
				if i == j {
					continue
				}
				for k := range deseasonalized {
					deseasonalized[k] -= seasonalValues[j][k]
				}
			}

			componentSeries := valuesToSeries(points, deseasonalized)
			result, err := STL(componentSeries, STLConfig{
				Period:          period,
				Seasonal:        cfg.SeasonalWindows[i],
				Robust:          cfg.Robust,
				InnerIterations: cfg.InnerIterations,
				OuterIterations: cfg.OuterIterations,
			})
			if err != nil {
				return MSTLResult{}, err
			}

			seasonalValues[i] = result.Seasonal.Values()
		}
	}

	totalSeasonal := make([]float64, len(values))
	for _, seasonal := range seasonalValues {
		for i := range totalSeasonal {
			totalSeasonal[i] += seasonal[i]
		}
	}

	deseasonalized := subtract(values, totalSeasonal)
	trendConfig, err := resolveSTLConfig(len(values), STLConfig{
		Period:          cfg.Periods[len(cfg.Periods)-1],
		InnerIterations: maxInt(cfg.InnerIterations, 1),
		Robust:          cfg.Robust,
		OuterIterations: cfg.OuterIterations,
	})
	if err != nil {
		return MSTLResult{}, err
	}

	trendWeights := ones(len(values))
	trend := loessSmooth(deseasonalized, trendConfig.Trend, trendWeights)
	if cfg.Robust {
		for i := 0; i < trendConfig.OuterIterations; i++ {
			residual := subtract(deseasonalized, trend)
			trendWeights = bisquareWeights(residual)
			trend = loessSmooth(deseasonalized, trendConfig.Trend, trendWeights)
		}
	}

	residual := subtract(deseasonalized, trend)

	components := make([]SeasonalComponent, 0, len(cfg.Periods))
	for i, period := range cfg.Periods {
		components = append(components, SeasonalComponent{
			Period: period,
			Series: valuesToSeries(points, seasonalValues[i]),
		})
	}

	return MSTLResult{
		Trend:    valuesToSeries(points, trend),
		Seasonal: components,
		Residual: valuesToSeries(points, residual),
	}, nil
}

func resolveMSTLConfig(seriesLength int, input MSTLConfig) (MSTLConfig, error) {
	if len(input.Periods) == 0 {
		return MSTLConfig{}, errors.New("at least one period is required")
	}
	if len(input.SeasonalWindows) > 0 && len(input.SeasonalWindows) != len(input.Periods) {
		return MSTLConfig{}, errors.New("seasonal windows must match periods length")
	}
	if input.Iterations < 0 {
		return MSTLConfig{}, errors.New("iterations cannot be negative")
	}

	cfg := input
	cfg.Periods = append([]int(nil), input.Periods...)
	sort.Ints(cfg.Periods)

	for i, period := range cfg.Periods {
		if period < 2 {
			return MSTLConfig{}, errors.New("periods must be at least 2")
		}
		if i > 0 && period == cfg.Periods[i-1] {
			return MSTLConfig{}, errors.New("periods must be unique")
		}
	}

	if seriesLength < 2*cfg.Periods[len(cfg.Periods)-1] {
		return MSTLConfig{}, errors.New("MSTL requires at least two complete cycles of the longest period")
	}

	if len(cfg.SeasonalWindows) == 0 {
		cfg.SeasonalWindows = make([]int, len(cfg.Periods))
	}
	for i, window := range cfg.SeasonalWindows {
		if window != 0 && (window < 3 || window%2 == 0) {
			return MSTLConfig{}, errors.New("seasonal windows must be odd and at least 3")
		}
		if window == 0 {
			cfg.SeasonalWindows[i] = 7
		}
	}

	return cfg, nil
}
