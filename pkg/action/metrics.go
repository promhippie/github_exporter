package action

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/promhippie/github_exporter/pkg/version"
)

var (
	registry  = prometheus.NewRegistry()
	namespace = "github"
)

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_seconds",
			Help:      "Histogram of latencies for requests to the api per collector.",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
		},
		[]string{"collector"},
	)

	requestFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "request_failures_total",
			Help:      "Total number of failed requests to the api per collector.",
		},
		[]string{"collector"},
	)
)

func init() {
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
		Namespace: namespace,
	}))

	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(version.Collector(namespace))

	registry.MustRegister(requestDuration)
	registry.MustRegister(requestFailures)
}

type promLogger struct {
	logger *slog.Logger
}

func (pl promLogger) Println(v ...interface{}) {
	pl.logger.Error(fmt.Sprintln(v...))
}
