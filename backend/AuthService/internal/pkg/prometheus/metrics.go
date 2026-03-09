package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics - структура метрик
type Metrics struct {
	requestError    *prometheus.CounterVec
	requestSuccess  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

// NewMetrics - конструктор
func NewMetrics(name string) *Metrics {
	return &Metrics{
		requestSuccess: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: name + "_success_rate_total",
			Help: "The total number of success requests",
		}, []string{"method"}),
		requestError: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: name + "_errors_rate_total",
			Help: "The total number of failed requests",
		}, []string{"method"}),
		requestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: name + "_duration",
			Help: "Request duration in seconds",
			Buckets: []float64{
				.005, .01, .025, .05, .1, .25, .5, 1, 2.5,
				5, 10, 20, 30, 40, 50, 60, 120, 180, 240, 300,
			},
		}, []string{"method"}),
	}
}

// HitSuccess - метрики для успехов
func (m *Metrics) HitSuccess(method string) {
	m.requestSuccess.WithLabelValues(method).Inc()
}

// HitError - метрики для ошибок
func (m *Metrics) HitError(method string) {
	m.requestError.WithLabelValues(method).Inc()
}

// HitDuration - метрики для времени
func (m *Metrics) HitDuration(method string, duration float64) {
	m.requestDuration.WithLabelValues(method).Observe(duration)
}
