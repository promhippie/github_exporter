package version

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Collector simply exports the version information for Prometheus.
func Collector(ns string) *prometheus.GaugeVec {
	info := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "build_info",
			Help:      "A metric with a constant '1' value labeled by version, revision and goversion from which it was built.",
		},
		[]string{"version", "revision", "goversion"},
	)

	info.WithLabelValues(String, Revision, Go).Set(1)
	return info
}
