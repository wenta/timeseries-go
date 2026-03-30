package plot

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	gonumplot "gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"

	timeseriesgo "github.com/wenta/timeseries-go"
)

// Style controls how a series is drawn on the chart.
type Style int

const (
	// Line draws a continuous line through all datapoints.
	Line Style = iota
	// Points draws only point markers without connecting lines.
	Points
	// LinePoints draws both a line and point markers.
	LinePoints
)

var (
	Gold           color.Color = color.RGBA{R: 234, G: 179, B: 8, A: 255}
	DeepSkyBlue    color.Color = color.RGBA{R: 14, G: 165, B: 233, A: 255}
	MediumPurple   color.Color = color.RGBA{R: 147, G: 51, B: 234, A: 255}
	Chartreuse     color.Color = color.RGBA{R: 132, G: 204, B: 22, A: 255}
	DarkOrange     color.Color = color.RGBA{R: 249, G: 115, B: 22, A: 255}
	Orchid         color.Color = color.RGBA{R: 217, G: 70, B: 239, A: 255}
	LightSeaGreen  color.Color = color.RGBA{R: 13, G: 148, B: 136, A: 255}
	Cyan           color.Color = color.RGBA{R: 6, G: 182, B: 212, A: 255}
	Crimson        color.Color = color.RGBA{R: 220, G: 38, B: 38, A: 255}
	LightBlue      color.Color = color.RGBA{R: 59, G: 130, B: 246, A: 255}
	DarkGray       color.Color = color.RGBA{R: 71, G: 85, B: 105, A: 255}
	SlateGray      color.Color = color.RGBA{R: 100, G: 116, B: 139, A: 255}
	LightSteelBlue color.Color = color.RGBA{R: 148, G: 163, B: 184, A: 255}
	White          color.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
)

var defaultPalette = []color.Color{
	Gold,
	DeepSkyBlue,
	MediumPurple,
	Chartreuse,
	DarkOrange,
	Orchid,
	LightSeaGreen,
	Cyan,
}

// Series describes one plotted timeseries together with its visual settings.
// Data is the input series, Label is used in the legend, Color overrides the
// default palette, and Style controls whether the series is drawn as a line,
// points, or both.
type Series struct {
	Label string
	Data  timeseriesgo.TimeSeries
	Color color.Color
	Style Style
}

// Option customizes chart rendering.
type Option func(*config)

type config struct {
	title      string
	xLabel     string
	yLabel     string
	widthPx    int
	heightPx   int
	timeFormat string
	showGrid   bool
	showLegend bool
	background color.Color
}

/**
 * Title sets the chart title shown above the plot.
 *
 * @param title Text shown as the chart title.
 * @return A plot.Option that applies the title.
 */
func Title(title string) Option {
	return func(c *config) {
		c.title = title
	}
}

/**
 * XLabel sets the horizontal axis label.
 *
 * @param label Text shown below the horizontal axis.
 * @return A plot.Option that applies the x-axis label.
 */
func XLabel(label string) Option {
	return func(c *config) {
		c.xLabel = label
	}
}

/**
 * YLabel sets the vertical axis label.
 *
 * @param label Text shown next to the vertical axis.
 * @return A plot.Option that applies the y-axis label.
 */
func YLabel(label string) Option {
	return func(c *config) {
		c.yLabel = label
	}
}

/**
 * Width sets the output width in pixels for HTML, SVG, and PNG rendering.
 *
 * @param width Output width in pixels.
 * @return A plot.Option that applies the width.
 */
func Width(width int) Option {
	return func(c *config) {
		if width > 0 {
			c.widthPx = width
		}
	}
}

/**
 * Height sets the output height in pixels for HTML, SVG, and PNG rendering.
 *
 * @param height Output height in pixels.
 * @return A plot.Option that applies the height.
 */
func Height(height int) Option {
	return func(c *config) {
		if height > 0 {
			c.heightPx = height
		}
	}
}

/**
 * TimeFormat sets the time layout used for x-axis tick labels.
 * If not provided, the renderer chooses a format automatically.
 *
 * @param layout Go time layout used to format x-axis ticks.
 * @return A plot.Option that applies the x-axis time format.
 */
func TimeFormat(layout string) Option {
	return func(c *config) {
		if layout != "" {
			c.timeFormat = layout
		}
	}
}

/**
 * WithoutGrid disables background grid lines.
 *
 * @return A plot.Option that disables grid rendering.
 */
func WithoutGrid() Option {
	return func(c *config) {
		c.showGrid = false
	}
}

/**
 * WithoutLegend hides the legend even when series labels are present.
 *
 * @return A plot.Option that disables legend rendering.
 */
func WithoutLegend() Option {
	return func(c *config) {
		c.showLegend = false
	}
}

/**
 * Background sets the chart background color.
 *
 * @param clr Background color used for the chart canvas.
 * @return A plot.Option that applies the background color.
 */
func Background(clr color.Color) Option {
	return func(c *config) {
		if clr != nil {
			c.background = clr
		}
	}
}

/**
 * Save renders series and writes the result to path.
 * The output format is selected from the file extension: .html, .svg, or .png.
 *
 * @param path Destination file path including extension.
 * @param series One or more plotted series to render.
 * @param opts Optional rendering options.
 * @return An error when rendering fails or the extension is unsupported.
 */
func Save(path string, series []Series, opts ...Option) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html":
		return SaveHTML(path, series, opts...)
	case ".svg":
		return SaveSVG(path, series, opts...)
	case ".png":
		return SavePNG(path, series, opts...)
	default:
		return fmt.Errorf("unsupported plot format %q", ext)
	}
}

/**
 * SaveSeries renders a single timeseries and writes it to path.
 * The output format is selected from the file extension.
 *
 * @param path Destination file path including extension.
 * @param ts TimeSeries to render.
 * @param opts Optional rendering options.
 * @return An error when rendering fails or the extension is unsupported.
 */
func SaveSeries(path string, ts timeseriesgo.TimeSeries, opts ...Option) error {
	return Save(path, []Series{{Data: ts}}, opts...)
}

/**
 * SaveHTML renders series as a standalone HTML document and writes it to path.
 *
 * @param path Destination .html file path.
 * @param series One or more plotted series to render.
 * @param opts Optional rendering options.
 * @return An error when rendering or writing fails.
 */
func SaveHTML(path string, series []Series, opts ...Option) error {
	return saveBytes(path, HTML, series, opts...)
}

/**
 * SaveSVG renders series as SVG markup and writes it to path.
 *
 * @param path Destination .svg file path.
 * @param series One or more plotted series to render.
 * @param opts Optional rendering options.
 * @return An error when rendering or writing fails.
 */
func SaveSVG(path string, series []Series, opts ...Option) error {
	return saveBytes(path, SVG, series, opts...)
}

/**
 * SavePNG renders series as a PNG image and writes it to path.
 *
 * @param path Destination .png file path.
 * @param series One or more plotted series to render.
 * @param opts Optional rendering options.
 * @return An error when rendering or writing fails.
 */
func SavePNG(path string, series []Series, opts ...Option) error {
	return saveBytes(path, PNG, series, opts...)
}

/**
 * SaveHTMLSeries renders a single timeseries as HTML and writes it to path.
 *
 * @param path Destination .html file path.
 * @param ts TimeSeries to render.
 * @param opts Optional rendering options.
 * @return An error when rendering or writing fails.
 */
func SaveHTMLSeries(path string, ts timeseriesgo.TimeSeries, opts ...Option) error {
	return SaveHTML(path, []Series{{Data: ts}}, opts...)
}

/**
 * SaveSVGSeries renders a single timeseries as SVG and writes it to path.
 *
 * @param path Destination .svg file path.
 * @param ts TimeSeries to render.
 * @param opts Optional rendering options.
 * @return An error when rendering or writing fails.
 */
func SaveSVGSeries(path string, ts timeseriesgo.TimeSeries, opts ...Option) error {
	return SaveSVG(path, []Series{{Data: ts}}, opts...)
}

/**
 * SavePNGSeries renders a single timeseries as PNG and writes it to path.
 *
 * @param path Destination .png file path.
 * @param ts TimeSeries to render.
 * @param opts Optional rendering options.
 * @return An error when rendering or writing fails.
 */
func SavePNGSeries(path string, ts timeseriesgo.TimeSeries, opts ...Option) error {
	return SavePNG(path, []Series{{Data: ts}}, opts...)
}

/**
 * HTML renders series as a standalone HTML document with embedded SVG.
 * The returned bytes can be written to disk or served directly over HTTP.
 *
 * @param series One or more plotted series to render.
 * @param opts Optional rendering options.
 * @return HTML bytes containing a complete document with the rendered chart.
 * @return An error when rendering fails.
 */
func HTML(series []Series, opts ...Option) ([]byte, error) {
	svgBytes, err := SVG(series, opts...)
	if err != nil {
		return nil, err
	}

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	svgText := stripXMLHeader(string(svgBytes))
	title := cfg.title
	if title == "" {
		title = "timeseries-go chart"
	}

	var out bytes.Buffer
	tpl := template.Must(template.New("chart").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    body {
      margin: 24px;
      font-family: ui-sans-serif, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: #f8fafc;
      color: #0f172a;
    }
    .chart {
      max-width: 100%;
      overflow-x: auto;
      background: #ffffff;
      border: 1px solid #e2e8f0;
      border-radius: 12px;
      padding: 16px;
      box-shadow: 0 8px 24px rgba(15, 23, 42, 0.06);
    }
    .chart svg {
      display: block;
      max-width: 100%;
      height: auto;
    }
  </style>
</head>
<body>
  <div class="chart">{{.SVG}}</div>
</body>
</html>
`))
	data := struct {
		Title string
		SVG   template.HTML
	}{
		Title: title,
		SVG:   template.HTML(svgText),
	}
	if err := tpl.Execute(&out, data); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

/**
 * SVG renders series as raw SVG bytes.
 *
 * @param series One or more plotted series to render.
 * @param opts Optional rendering options.
 * @return Raw SVG bytes for the rendered chart.
 * @return An error when rendering fails.
 */
func SVG(series []Series, opts ...Option) ([]byte, error) {
	return render(series, "svg", opts...)
}

/**
 * PNG renders series as PNG bytes.
 *
 * @param series One or more plotted series to render.
 * @param opts Optional rendering options.
 * @return Raw PNG bytes for the rendered chart.
 * @return An error when rendering fails.
 */
func PNG(series []Series, opts ...Option) ([]byte, error) {
	return render(series, "png", opts...)
}

func defaultConfig() config {
	return config{
		widthPx:    1280,
		heightPx:   480,
		showGrid:   true,
		showLegend: true,
		background: White,
	}
}

func render(series []Series, format string, opts ...Option) ([]byte, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	plt, err := buildPlot(series, cfg)
	if err != nil {
		return nil, err
	}

	writerTo, err := plt.WriterTo(pxToVG(cfg.widthPx), pxToVG(cfg.heightPx), format)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if _, err := writerTo.WriteTo(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func buildPlot(series []Series, cfg config) (*gonumplot.Plot, error) {
	filtered := make([]Series, 0, len(series))
	for _, s := range series {
		if !s.Data.IsEmpty() {
			filtered = append(filtered, s)
		}
	}
	if len(filtered) == 0 {
		return nil, errors.New("plot has no non-empty series")
	}

	plt := gonumplot.New()
	plt.Title.Text = cfg.title
	plt.X.Label.Text = cfg.xLabel
	plt.Y.Label.Text = cfg.yLabel
	plt.BackgroundColor = cfg.background
	plt.X.Tick.Marker = gonumplot.TimeTicks{
		Format: autoTimeFormat(filtered, cfg.timeFormat),
		Time:   gonumplot.UTCUnixTime,
	}

	if cfg.showGrid {
		grid := plotter.NewGrid()
		grid.Horizontal.Color = color.RGBA{R: 226, G: 232, B: 240, A: 255}
		grid.Vertical.Color = color.RGBA{R: 226, G: 232, B: 240, A: 255}
		plt.Add(grid)
	}

	for i, s := range filtered {
		clr := s.Color
		if clr == nil {
			clr = defaultPalette[i%len(defaultPalette)]
		}

		xys := toXYs(s.Data)
		style := s.Style

		switch style {
		case Points:
			scatter, err := plotter.NewScatter(xys)
			if err != nil {
				return nil, err
			}
			scatter.GlyphStyle.Color = clr
			scatter.GlyphStyle.Radius = vg.Points(2.5)
			scatter.GlyphStyle.Shape = draw.CircleGlyph{}
			plt.Add(scatter)
			if cfg.showLegend && s.Label != "" {
				plt.Legend.Add(s.Label, scatter)
			}
		case LinePoints:
			line, scatter, err := plotter.NewLinePoints(xys)
			if err != nil {
				return nil, err
			}
			line.LineStyle.Color = clr
			line.LineStyle.Width = vg.Points(1.6)
			scatter.GlyphStyle.Color = clr
			scatter.GlyphStyle.Radius = vg.Points(2.5)
			scatter.GlyphStyle.Shape = draw.CircleGlyph{}
			plt.Add(line, scatter)
			if cfg.showLegend && s.Label != "" {
				plt.Legend.Add(s.Label, line)
			}
		default:
			line, err := plotter.NewLine(xys)
			if err != nil {
				return nil, err
			}
			line.LineStyle.Color = clr
			line.LineStyle.Width = vg.Points(1.6)
			plt.Add(line)
			if cfg.showLegend && s.Label != "" {
				plt.Legend.Add(s.Label, line)
			}
		}
	}

	return plt, nil
}

func toXYs(ts timeseriesgo.TimeSeries) plotter.XYs {
	points := ts.DataPoints()
	xys := make(plotter.XYs, 0, len(points))
	for _, p := range points {
		if math.IsNaN(p.Value) || math.IsInf(p.Value, 0) {
			continue
		}
		xys = append(xys, plotter.XY{
			X: float64(p.Timestamp.UTC().Unix()),
			Y: p.Value,
		})
	}
	return xys
}

func autoTimeFormat(series []Series, explicit string) string {
	if explicit != "" {
		return explicit
	}

	start, end, ok := timeRange(series)
	if !ok {
		return time.RFC3339
	}

	span := end.Sub(start)
	switch {
	case span <= 48*time.Hour:
		return "Jan 02 15:04"
	case span <= 90*24*time.Hour:
		return "Jan 02"
	case span <= 3*365*24*time.Hour:
		return "Jan 2006"
	default:
		return "2006"
	}
}

func timeRange(series []Series) (time.Time, time.Time, bool) {
	var start time.Time
	var end time.Time
	ok := false
	for _, s := range series {
		points := s.Data.DataPoints()
		if len(points) == 0 {
			continue
		}
		first := points[0].Timestamp
		last := points[len(points)-1].Timestamp
		if !ok || first.Before(start) {
			start = first
		}
		if !ok || last.After(end) {
			end = last
		}
		ok = true
	}
	return start, end, ok
}

func pxToVG(px int) vg.Length {
	return vg.Inch * vg.Length(float64(px)/96.0)
}

func stripXMLHeader(svg string) string {
	trimmed := strings.TrimSpace(svg)
	if strings.HasPrefix(trimmed, "<?xml") {
		if idx := strings.Index(trimmed, "?>"); idx >= 0 {
			trimmed = strings.TrimSpace(trimmed[idx+2:])
		}
	}
	return trimmed
}

func saveBytes(path string, renderer func([]Series, ...Option) ([]byte, error), series []Series, opts ...Option) error {
	data, err := renderer(series, opts...)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
