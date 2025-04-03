//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	"github.com/promhippie/github_exporter/pkg/config"
)

func main() {
	f, err := os.Create("docs/partials/labels.md")

	if err != nil {
		fmt.Printf("failed to create file")
		os.Exit(1)
	}

	defer f.Close()

	f.WriteString("### Workflow Run Labels\n\n")
	for _, row := range config.RunLabels() {
		f.WriteString(fmt.Sprintf(
			"* %s\n",
			row,
		))
	}

	f.WriteString("\n### Workflow Job Labels\n\n")
	for _, row := range config.JobLabels() {
		f.WriteString(fmt.Sprintf(
			"* %s\n",
			row,
		))
	}

	f.WriteString("\n### Hosted Runner Labels\n\n")
	for _, row := range config.RunnerLabels() {
		f.WriteString(fmt.Sprintf(
			"* %s\n",
			row,
		))
	}
}
