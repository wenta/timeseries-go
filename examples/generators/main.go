package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/wenta/timeseries-go/generator"
	"github.com/wenta/timeseries-go/internal/exampleutil"
	"github.com/wenta/timeseries-go/plot"
)

func main() {
	outDir, err := exampleutil.OutputDir("generators")
	if err != nil {
		log.Fatalf("output dir creation failed: %v", err)
	}

	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	index := generator.MakeSeriesIndex(base, 30*time.Minute, 24)

	constant := generator.Constant(index, 5)
	walk := generator.RandomWalk(index, 10)
	noise := generator.RandomNoise(index, 0, 1)

	patternIndex := generator.MakeSeriesIndex(base, time.Hour, 4)
	pattern := generator.Constant(patternIndex, 2)
	repeated := generator.Repeat(pattern, base, base.Add(12*time.Hour))

	mustSave(outDir, "constant", "Constant Generator", []plot.Series{
		{Label: "constant", Data: constant, Color: plot.LightSeaGreen},
	})
	mustSave(outDir, "constant_vs_walk", "Constant vs Random Walk", []plot.Series{
		{Label: "constant", Data: constant, Color: plot.LightSeaGreen},
		{Label: "walk", Data: walk, Color: plot.MediumPurple},
	})
	mustSave(outDir, "constant_vs_noise", "Constant vs Random Noise", []plot.Series{
		{Label: "constant", Data: constant, Color: plot.LightSeaGreen},
		{Label: "noise", Data: noise, Color: plot.Orchid},
	})
	mustSave(outDir, "pattern_vs_repeat", "Pattern vs Repeat", []plot.Series{
		{Label: "pattern", Data: pattern, Color: plot.Cyan, Style: plot.Points},
		{Label: "repeat", Data: repeated, Color: plot.Gold},
	})

	exampleutil.PrintOutputDir(outDir)
}

func mustSave(outDir string, slug string, title string, series []plot.Series) {
	if err := exampleutil.SaveAllFormats(filepath.Join(outDir, slug), series, plot.Title(title)); err != nil {
		log.Fatalf("%s plot failed: %v", slug, err)
	}
}
