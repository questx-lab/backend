package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/questx-lab/backend/internal/common"
)

func NewHandler() http.Handler {
	registry := prometheus.NewRegistry()

	// default collectors
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	for _, counter := range common.PromCounters {
		registry.MustRegister(counter)
	}

	for _, histogram := range common.PromHistograms {
		registry.MustRegister(histogram)
	}

	promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	return promHandler
}
