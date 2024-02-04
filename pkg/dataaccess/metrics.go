package dataaccess

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// GCSLatency is the duration of GCS queries.
var GCSLatency = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "gcs_latency",
		Help: "Duration of GCS queries",
	},
	[]string{"query"},
)
