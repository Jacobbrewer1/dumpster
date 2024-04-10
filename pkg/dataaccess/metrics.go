package dataaccess

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// StorageLatency is the duration of GCS queries.
var StorageLatency = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "storage_latency",
		Help: "Duration of storage queries",
	},
	[]string{"query"},
)
