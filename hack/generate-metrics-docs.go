package main

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/exporter"
)

type Metric struct {
	Name   string
	Help   string
	Labels []string
}

func main() {
	failures := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dummy",
			Name:      "request_failures_total",
			Help:      "Total number of failed requests to the api per collector.",
		},
		[]string{"collector"},
	)

	collectors := make([]*prometheus.Desc, 0)

	collectors = append(
		collectors,
		exporter.NewOrgCollector(nil, nil, failures, nil, config.Load().Target).Metrics()...,
	)

	collectors = append(
		collectors,
		exporter.NewRepoCollector(nil, nil, failures, nil, config.Load().Target).Metrics()...,
	)

	collectors = append(
		collectors,
		exporter.NewActionCollector(nil, nil, failures, nil, config.Load().Target).Metrics()...,
	)

	collectors = append(
		collectors,
		exporter.NewPackageCollector(nil, nil, failures, nil, config.Load().Target).Metrics()...,
	)

	collectors = append(
		collectors,
		exporter.NewStorageCollector(nil, nil, failures, nil, config.Load().Target).Metrics()...,
	)

	metrics := make([]Metric, 0)

	metrics = append(metrics, Metric{
		Name:   "github_request_duration_seconds",
		Help:   "Histogram of latencies for requests to the api per collector",
		Labels: []string{"collector"},
	})

	metrics = append(metrics, Metric{
		Name:   "github_request_failures_total",
		Help:   "Total number of failed requests to the api per collector",
		Labels: []string{"collector"},
	})

	for _, desc := range collectors {
		metric := Metric{
			Name:   reflect.ValueOf(desc).Elem().FieldByName("fqName").String(),
			Help:   reflect.ValueOf(desc).Elem().FieldByName("help").String(),
			Labels: make([]string, 0),
		}

		labels := reflect.ValueOf(desc).Elem().FieldByName("variableLabels")

		for i := 0; i < labels.Len(); i++ {
			metric.Labels = append(metric.Labels, labels.Index(i).String())
		}

		metrics = append(metrics, metric)
	}

	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})

	f, err := os.Create("docs/partials/metrics.md")

	if err != nil {
		fmt.Printf("failed to create file")
		os.Exit(1)
	}

	defer f.Close()

	last := metrics[len(metrics)-1]
	for _, metric := range metrics {
		f.WriteString(fmt.Sprintf(
			"%s{%s}\n",
			metric.Name,
			strings.Join(
				metric.Labels,
				", ",
			),
		))

		f.WriteString(fmt.Sprintf(
			": %s\n",
			metric.Help,
		))

		if metric.Name != last.Name {
			f.WriteString("\n")
		}
	}
}
