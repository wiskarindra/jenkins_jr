// Package instrument is used to to anything related to instrumentation.
// By default, it uses Prometheus.
package instrument

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	once      sync.Once
	histogram *prometheus.HistogramVec
	counter   *prometheus.CounterVec
	gauge     *prometheus.GaugeVec
)

func init() {
	once.Do(func() {
		histogram = newHistogramVec()
		prometheus.MustRegister(histogram)

		counter = newCounterVec()
		prometheus.MustRegister(counter)

		gauge = newGaugeVec()
		prometheus.MustRegister(gauge)
	})
}

// Task is a struct that holds a job and its name.
type Task struct {
	Name string
	Job  func() error
}

// ObserveTask receives a task and run it with instrumentation.
// The instrument will record its latency using `ObserveLatency` method.
func ObserveTask(task Task) error {
	startTime := time.Now()
	err := task.Job()
	elapsedTime := time.Since(startTime).Seconds()

	if err != nil {
		ObserveLatency("packen-instrument-observe-task", task.Name, "fail", elapsedTime)
	} else {
		ObserveLatency("packen-instrument-observe-task", task.Name, "ok", elapsedTime)
	}

	return err
}

// Handler is an http.HandlerFunc to serve metrics endpoint.
func Handler(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

// ObserveLatency is a method to write metric, especially service latency, to prometheus using histogram type.
// The name of the metric is `service_latency_seconds`.
// It has three labels: `method`, `action`, and `status`.
//
// Label method is used to identify which method the metric belongs to,
// such as HTTP GET or HTTP POST.
// Label action is used to identify which action the metric belongs to,
// such as function name or HTTP path.
// Label status should only have one of these two values, "ok" or "fail".
// Status "ok" indicates that the function behaves well.
// Status "fail" indicates that there is an error occured during the process.
//
// The last parameter is `latency`.
// It indicates the running time of a function.
func ObserveLatency(method, action, status string, latency float64) {
	histogram.WithLabelValues(method, action, status).Observe(latency)
}

// IncrementByOne is a method to write metric to prometheus using counter type.
// The name of the metric is `service_entity_counter`.
// It has three labels: `service`, `entity`, and `status`.
//
// Label service is used to identify which service the metric belongs to.
// Label entity is used to identify which entity of the service the metric belongs to.
// Label status should only have one of these two values, "ok" or "fail".
// Status "ok" indicates that the function behaves well.
// Status "fail" indicates that there is an error occured during the process.
func IncrementByOne(service, entity, status string) {
	counter.WithLabelValues(service, entity, status).Inc()
}

// Gauge is a method to write metric to prometheus using gauge type.
// The name of the metric is `service_entity_gauge`.
// It has three labels: `service`, `entity`, and `status`.
//
// Label service is used to identify which service the metric belongs to.
// Label entity is used to identify which entity of the service the metric belongs to.
// Label status should only have one of these two values, "ok" or "fail".
// Status "ok" indicates that the function behaves well.
// Status "fail" indicates that there is an error occured during the process.
//
// The last parameter is `value`.
// It indicates the current value of the entity.
func Gauge(service, entity, status string, value float64) {
	gauge.WithLabelValues(service, entity, status).Set(value)
}

func newHistogramVec() *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "service_latency_seconds",
		Help: "service latency in seconds",
	}, []string{"method", "action", "status"})
}

func newCounterVec() *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "service_entity_counter",
		Help: "number of entity of service recorded based on its status",
	}, []string{"service", "entity", "status"})
}

func newGaugeVec() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "service_entity_gauge",
		Help: "arbitrary value of an entity of service recorded at one time",
	}, []string{"service", "entity", "status"})
}
