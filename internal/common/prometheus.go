package common

import "github.com/prometheus/client_golang/prometheus"

var (
	PromGauges = map[string]*prometheus.GaugeVec{}

	PromCounters = map[string]*prometheus.CounterVec{
		"http_requests_total": prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of all HTTP requests",
		}, []string{"method", "status_code"}),
	}

	PromHistograms = map[string]*prometheus.HistogramVec{
		"http_request_duration_seconds": prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of all HTTP requests",
		}, []string{"method", "status_code"}),
	}

	PromSummaries = map[string]*prometheus.SummaryVec{}
)
