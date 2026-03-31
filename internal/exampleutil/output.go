package exampleutil

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/plot"
	"github.com/wenta/timeseries-go/tsio"
)

/**
 * OutputDir creates and returns the output directory used by an example program.
 *
 * @param name Logical example name used as the output subdirectory.
 * @return The created directory path.
 * @return An error when the directory cannot be created.
 */
func OutputDir(name string) (string, error) {
	dir := filepath.Join("examples", "out", name)
	return dir, ensureDir(dir)
}

/**
 * SaveAllFormats renders series to HTML, SVG, and PNG files using basePath as prefix.
 *
 * @param basePath Path prefix used before the generated file extensions.
 * @param series One or more plotted series to render.
 * @param opts Optional rendering options passed to the plot package.
 * @return An error when any of the three renders fails.
 */
func SaveAllFormats(basePath string, series []plot.Series, opts ...plot.Option) error {
	for _, ext := range []string{".html", ".svg", ".png"} {
		if err := plot.Save(basePath+ext, series, opts...); err != nil {
			return err
		}
	}
	return nil
}

/**
 * SaveAllFormatsSeries renders a single series to HTML, SVG, and PNG files.
 *
 * @param basePath Path prefix used before the generated file extensions.
 * @param ts TimeSeries to render.
 * @param opts Optional rendering options passed to the plot package.
 * @return An error when any of the three renders fails.
 */
func SaveAllFormatsSeries(basePath string, ts timeseriesgo.TimeSeries, opts ...plot.Option) error {
	for _, ext := range []string{".html", ".svg", ".png"} {
		if err := plot.SaveSeries(basePath+ext, ts, opts...); err != nil {
			return err
		}
	}
	return nil
}

/**
 * FlaggedPoints filters base so that only points flagged by flags remain.
 * It is intended for plotting binary anomaly flags on top of the original series.
 *
 * @param base Original TimeSeries containing all datapoints.
 * @param flags Binary TimeSeries used to decide which datapoints to keep.
 * @return A TimeSeries containing only points from base where flags is greater than zero.
 */
func FlaggedPoints(base timeseriesgo.TimeSeries, flags timeseriesgo.TimeSeries) timeseriesgo.TimeSeries {
	if base.IsEmpty() || flags.IsEmpty() {
		return timeseriesgo.Empty()
	}

	flagByTime := make(map[int64]float64, flags.Length())
	for _, dp := range flags.DataPoints() {
		flagByTime[dp.Timestamp.UnixNano()] = dp.Value
	}

	points := make([]timeseriesgo.DataPoint, 0)
	for _, dp := range base.DataPoints() {
		if flagByTime[dp.Timestamp.UnixNano()] > 0 {
			points = append(points, dp)
		}
	}

	return timeseriesgo.FromDataPoints(points)
}

/**
 * AnchoredForecast prepends the last historical datapoint to a forecast series.
 * This makes forecast lines visually connect to the observed history on plots.
 *
 * @param history Original observed TimeSeries.
 * @param forecast Forecast TimeSeries containing future datapoints.
 * @return A TimeSeries that starts at the last point of history and continues with forecast.
 */
func AnchoredForecast(history timeseriesgo.TimeSeries, forecast timeseriesgo.TimeSeries) timeseriesgo.TimeSeries {
	if forecast.IsEmpty() {
		return forecast
	}

	last, err := history.Last()
	if err != nil {
		return forecast
	}

	points := make([]timeseriesgo.DataPoint, 0, forecast.Length()+1)
	forecastPoints := forecast.DataPoints()
	if len(forecastPoints) == 0 || !forecastPoints[0].Timestamp.Equal(last.Timestamp) {
		points = append(points, last)
	}
	points = append(points, forecastPoints...)
	return timeseriesgo.FromDataPoints(points)
}

/**
 * LoadCSVSeries reads a CSV file into a TimeSeries using the provided time layout.
 *
 * @param path File path to the CSV file.
 * @param timeFormat The Go time layout used to parse timestamps.
 * @param label The label to assign to the parsed TimeSeries.
 * @return A parsed TimeSeries or an error if the file cannot be read or parsed.
 */
func LoadCSVSeries(path string, timeFormat string, label string) (timeseriesgo.TimeSeries, error) {
	file, err := os.Open(path)
	if err != nil {
		return timeseriesgo.EmptyLabeled(label), err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return tsio.FromStringWithTimeFormat(*reader, timeFormat, label)
}

/**
 * PrintOutputDir prints the location where an example wrote its generated artifacts.
 *
 * @param dir Directory path containing generated example files.
 * @return None. The function writes a status line to stdout.
 */
func PrintOutputDir(dir string) {
	fmt.Printf("generated charts in %s\n", dir)
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}
