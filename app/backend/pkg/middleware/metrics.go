package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var defaultBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

type metricKey struct {
	service string
	method  string
	path    string
	status  string
}

type requestMetric struct {
	count   uint64
	sum     float64
	buckets []uint64
}

type metricStore struct {
	mu       sync.Mutex
	requests map[metricKey]*requestMetric
	inFlight map[string]int64
}

var globalMetrics = &metricStore{
	requests: make(map[metricKey]*requestMetric),
	inFlight: make(map[string]int64),
}

// HTTPMetrics records basic HTTP metrics in a Prometheus-compatible format.
// Count requests
func HTTPMetrics(service string) gin.HandlerFunc {
	service = normalizeService(service)

	return func(c *gin.Context) {
		start := time.Now()

		globalMetrics.addInFlight(service, 1)
		defer globalMetrics.addInFlight(service, -1)

		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		globalMetrics.observe(metricKey{
			service: service,
			method:  c.Request.Method,
			path:    path,
			status:  strconv.Itoa(c.Writer.Status()),
		}, time.Since(start).Seconds())
	}
}

// PrometheusHandler exposes collected metrics on /metrics.
// Expose metrics on /metrics
func PrometheusHandler(service string) gin.HandlerFunc {
	service = normalizeService(service)

	return func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(globalMetrics.render(service)))
	}
}

func (s *metricStore) addInFlight(service string, delta int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inFlight[service] += delta
}

func (s *metricStore) observe(key metricKey, seconds float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metric := s.requests[key]
	if metric == nil {
		metric = &requestMetric{buckets: make([]uint64, len(defaultBuckets))}
		s.requests[key] = metric
	}

	metric.count++
	metric.sum += seconds
	for i, bucket := range defaultBuckets {
		if seconds <= bucket {
			metric.buckets[i]++
		}
	}
}

func (s *metricStore) render(service string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	var b strings.Builder
	b.WriteString("# HELP http_requests_total Total number of HTTP requests.\n")
	b.WriteString("# TYPE http_requests_total counter\n")

	keys := make([]metricKey, 0, len(s.requests))
	for key := range s.requests {
		if key.service == service {
			keys = append(keys, key)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].path != keys[j].path {
			return keys[i].path < keys[j].path
		}
		if keys[i].method != keys[j].method {
			return keys[i].method < keys[j].method
		}
		return keys[i].status < keys[j].status
	})

	for _, key := range keys {
		metric := s.requests[key]
		fmt.Fprintf(&b, "http_requests_total{%s} %d\n", labels(key), metric.count)
	}

	b.WriteString("# HELP http_request_duration_seconds HTTP request latency in seconds.\n")
	b.WriteString("# TYPE http_request_duration_seconds histogram\n")
	for _, key := range keys {
		metric := s.requests[key]
		for i, bucket := range defaultBuckets {
			fmt.Fprintf(&b, "http_request_duration_seconds_bucket{%s,le=%q} %d\n", labels(key), formatFloat(bucket), metric.buckets[i])
		}
		fmt.Fprintf(&b, "http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", labels(key), metric.count)
		fmt.Fprintf(&b, "http_request_duration_seconds_sum{%s} %s\n", labels(key), formatFloat(metric.sum))
		fmt.Fprintf(&b, "http_request_duration_seconds_count{%s} %d\n", labels(key), metric.count)
	}

	b.WriteString("# HELP http_requests_in_flight Current number of HTTP requests being served.\n")
	b.WriteString("# TYPE http_requests_in_flight gauge\n")
	fmt.Fprintf(&b, "http_requests_in_flight{service=\"%s\"} %d\n", escapeLabel(service), s.inFlight[service])

	b.WriteString("# HELP service_runtime_goroutines Current number of goroutines.\n")
	b.WriteString("# TYPE service_runtime_goroutines gauge\n")
	fmt.Fprintf(&b, "service_runtime_goroutines{service=\"%s\"} %d\n", escapeLabel(service), runtime.NumGoroutine())

	b.WriteString("# HELP service_memory_alloc_bytes Bytes of allocated heap objects.\n")
	b.WriteString("# TYPE service_memory_alloc_bytes gauge\n")
	fmt.Fprintf(&b, "service_memory_alloc_bytes{service=\"%s\"} %d\n", escapeLabel(service), mem.Alloc)

	return b.String()
}

func labels(key metricKey) string {
	return fmt.Sprintf(
		"service=\"%s\",method=\"%s\",path=\"%s\",status=\"%s\"",
		escapeLabel(key.service),
		escapeLabel(key.method),
		escapeLabel(key.path),
		escapeLabel(key.status),
	)
}

func normalizeService(service string) string {
	service = strings.TrimSpace(service)
	if service == "" {
		return "unknown"
	}
	return service
}

func escapeLabel(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return strings.ReplaceAll(value, "\"", "\\\"")
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}
