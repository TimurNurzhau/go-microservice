package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	TotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func Init() {
	prometheus.MustRegister(TotalRequests)
	prometheus.MustRegister(RequestDuration)
}

func Instrument(path string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		TotalRequests.WithLabelValues(r.Method, path).Inc()
		h(w, r)
		RequestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	}
}

func Handler() http.Handler {
	return promhttp.Handler()
}
