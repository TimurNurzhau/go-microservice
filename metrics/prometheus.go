package metrics

import (
	"net/http"
	"strconv"
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

	HTTPErrorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP responses with status code 4xx or 5xx",
		},
		[]string{"method", "endpoint", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func Init() {
	prometheus.MustRegister(
		TotalRequests,
		HTTPErrorTotal,
		RequestDuration,
	)
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusRecorder) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}

	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(body)
}

func (w *statusRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Instrument binds metric children when the route is registered, rather than
// looking them up by method and endpoint labels for every request.
func Instrument(method, endpoint string, next http.Handler) http.Handler {
	requestCounter := TotalRequests.WithLabelValues(method, endpoint)
	durationObserver := RequestDuration.WithLabelValues(method, endpoint)

	errorCounters := make(map[int]prometheus.Counter)
	for _, code := range []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
		http.StatusRequestTimeout,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	} {
		errorCounters[code] = HTTPErrorTotal.WithLabelValues(
			method,
			endpoint,
			http.StatusText(code),
		)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		requestCounter.Inc()
		durationObserver.Observe(time.Since(start).Seconds())

		if recorder.status >= http.StatusBadRequest {
			if counter, ok := errorCounters[recorder.status]; ok {
				counter.Inc()
			} else {
				HTTPErrorTotal.WithLabelValues(
					method,
					endpoint,
					strconv.Itoa(recorder.status),
				).Inc()
			}
		}
	})
}

func Handler() http.Handler {
	return promhttp.Handler()
}
