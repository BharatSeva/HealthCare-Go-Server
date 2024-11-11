package main
import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Total requests counter
	totalRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Request duration histogram
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Response size histogram
	responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes.",
			Buckets: []float64{100, 500, 1000, 5000, 10000},
		},
		[]string{"method", "endpoint"},
	)

	// Active requests gauge
	activeRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_active_requests",
		Help: "Number of active HTTP requests.",
	})

	// Failed requests counter
	failedRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_failed_total",
			Help: "Total number of failed HTTP requests.",
		},
		[]string{"method", "endpoint", "error"},
	)
)

// PrometheusMiddleware implements mux.MiddlewareFunc
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		activeRequests.Inc()

		// Create response writer wrapper to capture status code
		rw := NewResponseWriter(w)

		next.ServeHTTP(rw, r)

		// Decrease active requests counter
		activeRequests.Dec()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := rw.statusCode

		// Update request total
		totalRequests.WithLabelValues(r.Method, r.URL.Path, string(rune(status))).Inc()

		// Update duration histogram
		requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)

		// Update response size
		responseSize.WithLabelValues(r.Method, r.URL.Path).Observe(float64(rw.size))

		// Record failed requests (status >= 400)
		if status >= 400 {
			failedRequests.WithLabelValues(r.Method, r.URL.Path, string(rune(status))).Inc()
		}
	})
}

// ResponseWriter wrapper to capture status code and response size
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: w}
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
