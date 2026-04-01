package exampleutil

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"

	timeseriesgo "github.com/wenta/timeseries-go"
	"github.com/wenta/timeseries-go/plot"
	"github.com/wenta/timeseries-go/tsio"
)

type Report struct {
	name   string
	title  string
	charts []reportChart
}

type reportChart struct {
	Slug  string
	Title string
	SVG   template.HTML
}

type examplesIndexEntry struct {
	Name  string
	Title string
	Path  string
}

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
 * NewReport creates a chart report collector for one example program.
 *
 * @param name Logical example name used in the output index.
 * @param title Human-readable report title shown in the generated HTML.
 * @return A Report that can collect charts and write a single HTML report.
 */
func NewReport(name string, title string) *Report {
	return &Report{name: name, title: title}
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

/**
 * SaveChart stores SVG and PNG assets for one chart and adds it to the HTML report.
 *
 * @param outDir Output directory for the example program.
 * @param slug File name prefix used for generated assets.
 * @param title Chart title shown in the report.
 * @param series One or more plotted series to render.
 * @param opts Optional rendering options passed to the plot package.
 * @return An error when asset rendering fails.
 */
func (r *Report) SaveChart(outDir string, slug string, title string, series []plot.Series, opts ...plot.Option) error {
	basePath := filepath.Join(outDir, slug)
	renderOpts := append([]plot.Option{plot.Title(title)}, opts...)

	for _, ext := range []string{".svg", ".png"} {
		if err := plot.Save(basePath+ext, series, renderOpts...); err != nil {
			return err
		}
	}

	svgBytes, err := plot.SVG(series, renderOpts...)
	if err != nil {
		return err
	}

	r.charts = append(r.charts, reportChart{
		Slug:  slug,
		Title: title,
		SVG:   template.HTML(string(svgBytes)),
	})
	return nil
}

/**
 * SaveChartSeries stores SVG and PNG assets for one single-series chart and adds it to the HTML report.
 *
 * @param outDir Output directory for the example program.
 * @param slug File name prefix used for generated assets.
 * @param title Chart title shown in the report.
 * @param ts TimeSeries to render.
 * @param opts Optional rendering options passed to the plot package.
 * @return An error when asset rendering fails.
 */
func (r *Report) SaveChartSeries(outDir string, slug string, title string, ts timeseriesgo.TimeSeries, opts ...plot.Option) error {
	return r.SaveChart(outDir, slug, title, []plot.Series{{Data: ts}}, opts...)
}

/**
 * Write stores the collected charts as a single HTML report and refreshes the top-level examples index.
 *
 * @param outDir Output directory for the example program.
 * @return The written report path.
 * @return An error when report writing fails.
 */
func (r *Report) Write(outDir string) (string, error) {
	reportPath := filepath.Join(outDir, "index.html")
	if err := writeExampleReport(reportPath, r.title, r.charts); err != nil {
		return "", err
	}
	if err := writeExamplesIndex(filepath.Join("examples", "out")); err != nil {
		return "", err
	}
	return reportPath, nil
}

func writeExampleReport(path string, title string, charts []reportChart) error {
	const page = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    :root { color-scheme: light; }
    body { margin: 0; padding: 32px; font-family: ui-sans-serif, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f8fafc; color: #0f172a; }
    main { max-width: 1400px; margin: 0 auto; }
    h1 { margin: 0 0 12px; font-size: 2rem; }
    p { margin: 0 0 24px; color: #475569; }
    nav { margin: 0 0 32px; display: flex; flex-wrap: wrap; gap: 12px; }
    nav a { text-decoration: none; color: #0369a1; background: #e0f2fe; padding: 8px 12px; border-radius: 999px; }
    section { margin: 0 0 28px; padding: 20px; background: white; border: 1px solid #e2e8f0; border-radius: 16px; box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06); }
    h2 { margin: 0 0 16px; font-size: 1.25rem; }
    .chart svg { width: 100%; height: auto; display: block; }
    .assets { margin-top: 12px; display: flex; gap: 12px; }
    .assets a { color: #0369a1; text-decoration: none; }
  </style>
</head>
<body>
  <main>
    <h1>{{.Title}}</h1>
    <p>Static SVG/PNG assets are saved next to this report.</p>
    <nav>
      {{range .Charts}}<a href="#{{.Slug}}">{{.Title}}</a>{{end}}
    </nav>
    {{range .Charts}}
    <section id="{{.Slug}}">
      <h2>{{.Title}}</h2>
      <div class="chart">{{.SVG}}</div>
      <div class="assets">
        <a href="./{{.Slug}}.svg">SVG</a>
        <a href="./{{.Slug}}.png">PNG</a>
      </div>
    </section>
    {{end}}
  </main>
</body>
</html>`

	tmpl, err := template.New("example-report").Parse(page)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data := struct {
		Title  string
		Charts []reportChart
	}{
		Title:  title,
		Charts: charts,
	}
	return tmpl.Execute(file, data)
}

func writeExamplesIndex(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	reports := make([]examplesIndexEntry, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		reportPath := filepath.Join(root, entry.Name(), "index.html")
		if _, err := os.Stat(reportPath); err != nil {
			continue
		}

		reports = append(reports, examplesIndexEntry{
			Name:  entry.Name(),
			Title: humanizeExampleName(entry.Name()),
			Path:  filepath.ToSlash(filepath.Join(entry.Name(), "index.html")),
		})
	}

	sort.Slice(reports, func(i int, j int) bool {
		return reports[i].Name < reports[j].Name
	})

	const page = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>timeseries-go examples</title>
  <style>
    body { margin: 0; padding: 32px; font-family: ui-sans-serif, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f8fafc; color: #0f172a; }
    main { max-width: 960px; margin: 0 auto; }
    h1 { margin: 0 0 12px; font-size: 2rem; }
    p { margin: 0 0 24px; color: #475569; }
    ul { list-style: none; padding: 0; margin: 0; display: grid; gap: 12px; }
    li { background: white; border: 1px solid #e2e8f0; border-radius: 14px; }
    a { display: block; padding: 18px 20px; color: #0369a1; text-decoration: none; }
  </style>
</head>
<body>
  <main>
    <h1>timeseries-go examples</h1>
    <p>Collected HTML reports generated by the example programs.</p>
    <ul>
      {{range .Reports}}<li><a href="./{{.Path}}">{{.Title}}</a></li>{{end}}
    </ul>
  </main>
</body>
</html>`

	tmpl, err := template.New("examples-index").Parse(page)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(root, "index.html"))
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, struct {
		Reports []examplesIndexEntry
	}{Reports: reports})
}

func humanizeExampleName(name string) string {
	switch name {
	case "main":
		return "Guided Tour"
	default:
		return name
	}
}
