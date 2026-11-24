package metrics

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
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

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Используем шаблон маршрута (/api/users/{id}), а не конкретный URL с ID
		route := mux.CurrentRoute(r)
		path := "/unknown"
		if route != nil {
			if tmpl, err := route.GetPathTemplate(); err == nil {
				path = tmpl
			} else {
				path = r.URL.Path
			}
		} else {
			path = r.URL.Path
		}

		TotalRequests.WithLabelValues(r.Method, path).Inc()
		next.ServeHTTP(w, r)
		RequestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	})
}

func Handler() http.Handler {
	return promhttp.Handler()
}