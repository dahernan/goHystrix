package metrics

import (
	"sync"
	"time"
)

var (
	metrics = NewMetricsHolder()
)

type MetricsHolder struct {
	metrics map[string]map[string]*Metric
	mutex   sync.RWMutex
}

type Metric struct {
	name  string
	group string

	failures       Counter
	success        Counter
	fallback       Counter
	fallbackErrors Counter
	timeouts       Counter

	consecutiveFailures Counter
	consecutiveSuccess  Counter
	consecutiveTimeouts Counter

	lastFailure time.Time
	lastSuccess time.Time
	lastTimeout time.Time
}

func (m *Metric) Success() {
	// TODO: rethink !! is not concurrency safe
	m.success.Inc(1)
	m.consecutiveSuccess.Inc(1)
	m.consecutiveFailures.Clear()
	m.lastSuccess = time.Now()
}

func (m *Metric) Fail() {
	m.failures.Inc(1)
	m.consecutiveSuccess.Clear()
	m.consecutiveFailures.Inc(1)
	m.lastFailure = time.Now()
}

func (m *Metric) Fallback() {
	m.fallback.Inc(1)
}

func (m *Metric) FallbackError() {
	m.fallbackErrors.Inc(1)
}

func (m *Metric) Timeout() {
	m.timeouts.Inc(1)
	m.failures.Inc(1)
	m.consecutiveSuccess.Clear()
	m.consecutiveFailures.Inc(1)
	m.lastFailure = time.Now()
	m.lastTimeout = time.Now()
}

func (m *Metric) FailuresCount() int64 {
	return m.failures.Count()
}

func (m *Metric) SuccessCount() int64 {
	return m.success.Count()
}

func (m *Metric) TimeoutsCount() int64 {
	return m.timeouts.Count()
}

func (m *Metric) ConsecutiveFailures() int64 {
	return m.consecutiveFailures.Count()
}

func (m *Metric) LastFailure() time.Time {
	return m.lastFailure
}

func NewMetricsHolder() *MetricsHolder {
	return &MetricsHolder{metrics: make(map[string]map[string]*Metric)}
}

func Metrics() *MetricsHolder {
	return metrics
}
func MetricsReset() {
	metrics = NewMetricsHolder()
}

func NewMetric(group string, name string) *Metric {
	m := &Metric{}
	m.name = name
	m.group = group

	m.success = NewCounter()
	m.failures = NewCounter()
	m.fallback = NewCounter()
	m.fallbackErrors = NewCounter()
	m.timeouts = NewCounter()

	m.consecutiveFailures = NewCounter()
	m.consecutiveSuccess = NewCounter()
	m.consecutiveTimeouts = NewCounter()

	Metrics().Set(group, name, m)
	return m

}

func (holder *MetricsHolder) Get(group string, name string) (*Metric, bool) {
	holder.mutex.RLock()
	defer holder.mutex.RUnlock()
	metricValues, ok := holder.metrics[group]
	if !ok {
		return nil, ok
	}

	value, ok := metricValues[name]
	return value, ok
}

func (holder *MetricsHolder) Set(group string, name string, value *Metric) {
	holder.mutex.Lock()
	defer holder.mutex.Unlock()

	metricValues, ok := holder.metrics[group]
	if !ok {
		metricValues = make(map[string]*Metric)
		holder.metrics[group] = metricValues
	}

	metricValues[name] = value

}
