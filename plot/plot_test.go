package plot

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

func TestSVGIncludesTitle(t *testing.T) {
	ts := testSeries()

	data, err := SVG([]Series{{Label: "cpu", Data: ts}}, Title("CPU Load"))
	if err != nil {
		t.Fatalf("unexpected svg error: %v", err)
	}

	out := string(data)
	if !strings.Contains(out, "<svg") {
		t.Fatalf("expected svg markup, got %q", out)
	}
	if !strings.Contains(out, "CPU Load") {
		t.Fatalf("expected plot title in svg, got %q", out)
	}
}

func TestHTMLEmbedsSVG(t *testing.T) {
	ts := testSeries()

	data, err := HTML([]Series{{Label: "cpu", Data: ts}}, Title("CPU Load"))
	if err != nil {
		t.Fatalf("unexpected html error: %v", err)
	}

	out := string(data)
	if !strings.Contains(out, "<!doctype html>") {
		t.Fatalf("expected html document, got %q", out)
	}
	if !strings.Contains(out, "<svg") {
		t.Fatalf("expected embedded svg, got %q", out)
	}
}

func TestPNGStartsWithMagicBytes(t *testing.T) {
	ts := testSeries()

	data, err := PNG([]Series{{Label: "cpu", Data: ts}}, Title("CPU Load"))
	if err != nil {
		t.Fatalf("unexpected png error: %v", err)
	}

	if !bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		t.Fatalf("expected png bytes, got %v", data[:8])
	}
}

func TestSaveInfersFormatFromExtension(t *testing.T) {
	ts := testSeries()
	dir := t.TempDir()

	for _, ext := range []string{".html", ".svg", ".png"} {
		path := filepath.Join(dir, "chart"+ext)
		if err := Save(path, []Series{{Label: "cpu", Data: ts}}, Title("CPU")); err != nil {
			t.Fatalf("save %s failed: %v", ext, err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s failed: %v", ext, err)
		}
		if info.Size() == 0 {
			t.Fatalf("expected non-empty file for %s", ext)
		}
	}
}

func TestSaveSeriesWritesSingleSeriesChart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "single.svg")
	if err := SaveSeries(path, testSeries(), Title("Single")); err != nil {
		t.Fatalf("save series failed: %v", err)
	}
}

func TestFormatSpecificSaveHelpers(t *testing.T) {
	ts := testSeries()
	dir := t.TempDir()

	if err := SaveHTML(filepath.Join(dir, "chart.html"), []Series{{Data: ts}}, Title("HTML")); err != nil {
		t.Fatalf("save html failed: %v", err)
	}
	if err := SaveSVGSeries(filepath.Join(dir, "chart.svg"), ts, Title("SVG")); err != nil {
		t.Fatalf("save svg series failed: %v", err)
	}
	if err := SavePNGSeries(filepath.Join(dir, "chart.png"), ts, Title("PNG")); err != nil {
		t.Fatalf("save png series failed: %v", err)
	}
}

func testSeries() timeseriesgo.TimeSeries {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	return timeseriesgo.FromDataPoints([]timeseriesgo.DataPoint{
		{Timestamp: start, Value: 1},
		{Timestamp: start.Add(time.Hour), Value: 2},
		{Timestamp: start.Add(2 * time.Hour), Value: 3},
		{Timestamp: start.Add(3 * time.Hour), Value: 2.5},
	})
}
