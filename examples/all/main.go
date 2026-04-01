package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	targets := []string{
		"./examples",
		"./examples/plotting",
		"./examples/forecasting",
		"./examples/anomalies",
		"./examples/decomposition",
		"./examples/transformations",
		"./examples/generators",
	}

	for _, target := range targets {
		fmt.Printf("running %s\n", target)

		cmd := exec.Command("go", "run", target)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("%s failed: %v", target, err)
		}
	}

	fmt.Println("all example reports generated in examples/out")
}
