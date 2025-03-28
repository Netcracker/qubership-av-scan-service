package clamav

import (
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
)

// DatabaseAgeMetric is the name of the metric which tracks
// the age of ClamAV database in seconds
const DatabaseAgeMetric = "av_database_age_seconds"

// Collector is used to collect ClamAV metrics for prometheus client
type Collector struct {
	client      Clamd
	log         *slog.Logger
	databaseAge *prometheus.Desc
}

// NewMetricsCollector creates a new Collector which
// collects metrics from given Clamd instance
func NewMetricsCollector(client Clamd, log *slog.Logger) *Collector {
	return &Collector{
		client: client,
		log:    log,
		databaseAge: prometheus.NewDesc(
			DatabaseAgeMetric,
			"Shows ClamAV viruses database age in seconds",
			nil,
			nil,
		),
	}
}

func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.databaseAge
}

func (collector *Collector) Collect(ch chan<- prometheus.Metric) {
	ageSeconds, err := collector.client.DatabaseAge()

	if err != nil {
		collector.log.Error("Failed to collect ClamAV DB age metric", "error", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		collector.databaseAge,
		prometheus.GaugeValue,
		ageSeconds,
	)
}
