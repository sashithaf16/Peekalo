package metrics

import "github.com/prometheus/client_golang/prometheus"

var PrometheusRegistry = prometheus.NewRegistry()

var (
	RequestInvalidCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "request_invalid_count",
			Help: "Number of invalid requests received",
		})

	RequestReceivedSuccessCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "request_received_success_count",
			Help: "Number of successfully received requests",
		})

	RequestAnalyzerSuccessCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "request_analyzer_success_count",
			Help: "Number of requests successfully analyzed",
		})

	RequestAnalyzerFailureCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "request_analyzer_failure_count",
			Help: "Number of requests failed to analyze",
		})
)

func RegisterMetrics() {
	PrometheusRegistry.MustRegister(RequestInvalidCount)
	PrometheusRegistry.MustRegister(RequestReceivedSuccessCount)
	PrometheusRegistry.MustRegister(RequestAnalyzerSuccessCount)
	PrometheusRegistry.MustRegister(RequestAnalyzerFailureCount)
}
