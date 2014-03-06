package metrics

import (
	"sync"
)

var (
	metrics = NewMetricsHolder()
)

type MetricsHolder struct {
	metrics map[string]map[string]*Metric
	mutex   sync.RWMutex
}

type Metric struct {
	name           string
	group          string
	Failures       Counter
	Success        Counter
	Fallback       Counter
	FallbackErrors Counter
	Timeouts       Counter
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

	m.Success = NewCounter()
	m.Failures = NewCounter()
	m.Fallback = NewCounter()
	m.FallbackErrors = NewCounter()
	m.Timeouts = NewCounter()

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
