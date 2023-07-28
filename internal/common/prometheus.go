package common

import "github.com/prometheus/client_golang/prometheus"

const (
	HTTPRequestTotal             = "http_requests_total"
	BlockchainTransactionFailure = "blockchain_transaction_failure"
	HTTPRequestDurationSeconds   = "http_request_duration_seconds"
)

var (
	PromGauges = map[string]*prometheus.GaugeVec{}

	PromCounters = map[string]*prometheus.CounterVec{
		HTTPRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: HTTPRequestTotal,
			Help: "Count of all HTTP requests",
		}, []string{"method", "status_code"}),
		BlockchainTransactionFailure: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: BlockchainTransactionFailure,
			Help: "Count of all blockchain transaction failure",
		}, []string{"method"}),
	}

	PromHistograms = map[string]*prometheus.HistogramVec{
		HTTPRequestDurationSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: HTTPRequestDurationSeconds,
			Help: "Duration of all HTTP requests",
		}, []string{"method", "status_code"}),
	}

	PromSummaries = map[string]*prometheus.SummaryVec{}
)
