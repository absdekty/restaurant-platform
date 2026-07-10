package middleware

import (
	"net/http"
	"sync"
	"sync/atomic"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

type Metrics struct {
	totalRequests  uint64
	activeRequests int64
	errorsTotal    uint64
	errorsByStatus map[string]*uint64
	mu             sync.RWMutex
}

func NewMetrics() *Metrics {
	return &Metrics{errorsByStatus: make(map[string]*uint64)}
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&m.totalRequests, 1)
		atomic.AddInt64(&m.activeRequests, 1)
		defer atomic.AddInt64(&m.activeRequests, -1)

		rr := &responseRecorder{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(rr, r)

		if rr.statusCode >= 400 {
			atomic.AddUint64(&m.errorsTotal, 1)
			m.incErrorsByStatus(rr.statusCode)
		}
	})
}

func (m *Metrics) GetTotalRequests() uint64 {
	return atomic.LoadUint64(&m.totalRequests)
}

func (m *Metrics) GetActiveRequests() int64 {
	return atomic.LoadInt64(&m.activeRequests)
}

func (m *Metrics) GetErrorsTotal() uint64 {
	return atomic.LoadUint64(&m.errorsTotal)
}

func (m *Metrics) GetErrorsByStatus() map[string]uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]uint64)
	for k, v := range m.errorsByStatus {
		result[k] = atomic.LoadUint64(v)
	}
	return result
}

func (m *Metrics) incErrorsByStatus(status int) {
	statusStr := http.StatusText(status)

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.errorsByStatus[statusStr]; !ok {
		m.errorsByStatus[statusStr] = new(uint64)
	}
	atomic.AddUint64(m.errorsByStatus[statusStr], 1)
}
