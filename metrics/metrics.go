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

	successChan       chan time.Duration
	failuresChan      chan struct{}
	fallbackChan      chan struct{}
	fallbackErrorChan chan struct{}
	timeoutsChan      chan struct{}
	countersChan      chan struct{}
	countersOutChan   chan HealthCounts

	buckets int
	window  time.Duration
	values  []HealthCountsBucket

	lastFailure time.Time
	lastSuccess time.Time
	lastTimeout time.Time
}

func NewMetric(group string, name string) *Metric {
	return NewMetricWithDuration(group, name, 20, 20*time.Second)
}

func NewMetricWithDuration(group string, name string, windowSize int, duration time.Duration) *Metric {
	m := &Metric{}

	m.name = name
	m.group = group

	m.buckets = windowSize
	m.window = duration
	m.values = make([]HealthCountsBucket, m.buckets, m.buckets)

	m.successChan = make(chan time.Duration)
	m.failuresChan = make(chan struct{})
	m.fallbackChan = make(chan struct{})
	m.fallbackErrorChan = make(chan struct{})
	m.timeoutsChan = make(chan struct{})
	m.countersChan = make(chan struct{})
	m.countersOutChan = make(chan HealthCounts)

	Metrics().Set(group, name, m)
	go m.run()
	return m

}

type HealthCountsBucket struct {
	Failures       int64
	Success        int64
	Fallback       int64
	FallbackErrors int64
	Timeouts       int64
	lastWrite      time.Time
}

type HealthCounts struct {
	HealthCountsBucket
	Total           int64
	ErrorPercentage float64
}

func (c *HealthCountsBucket) Reset() {
	c.Failures = 0
	c.Success = 0
	c.Fallback = 0
	c.FallbackErrors = 0
	c.Timeouts = 0
}

func (m *Metric) run() {
	for {
		select {
		case duration := <-m.successChan:
			m.doSuccess(duration)
		case <-m.failuresChan:
			m.doFail()
		case <-m.timeoutsChan:
			m.doTimeout()
		case <-m.fallbackChan:
			m.doFallback()
		case <-m.fallbackErrorChan:
			m.doFallbackError()
		case <-m.countersChan:
			m.countersOutChan <- m.doHealthCounts()
			//case <-time.After(2 * time.Second):
			//	fmt.Println("NOTHING :(")
		}
	}

}

func (m *Metric) bucket() *HealthCountsBucket {
	now := time.Now()
	index := now.Second() % m.buckets
	if !m.values[index].lastWrite.IsZero() {
		elapsed := now.Sub(m.values[index].lastWrite)
		if elapsed > m.window {
			m.values[index].Reset()
		}
	}
	m.values[index].lastWrite = now
	return &m.values[index]
}

func (m *Metric) doSuccess(duration time.Duration) {
	m.bucket().Success++
	m.lastSuccess = time.Now()
}

func (m *Metric) doFail() {
	m.bucket().Failures++
	m.lastFailure = time.Now()
}

func (m *Metric) doFallback() {
	m.bucket().Fallback++
}

func (m *Metric) doTimeout() {
	m.bucket().Timeouts++
	m.bucket().Failures++
	now := time.Now()
	m.lastFailure = now
	m.lastTimeout = now
}

func (m *Metric) doFallbackError() {
	m.bucket().FallbackErrors++
}

func (m *Metric) doHealthCounts() (counters HealthCounts) {
	now := time.Now()
	for _, value := range m.values {
		if !value.lastWrite.IsZero() && (now.Sub(value.lastWrite) <= m.window) {
			counters.Success += value.Success
			counters.Failures += value.Failures
			counters.Fallback += value.Fallback
			counters.FallbackErrors += value.FallbackErrors
			counters.Timeouts += value.Timeouts
		}
	}
	counters.Total = counters.Success + counters.Failures
	if counters.Total == 0 {
		counters.ErrorPercentage = 0
	} else {
		counters.ErrorPercentage = float64(counters.Failures) / float64(counters.Total) * 100.0
	}
	return
}

func (m *Metric) HealthCounts() HealthCounts {
	m.countersChan <- struct{}{}
	return <-m.countersOutChan
}

func (m *Metric) Success(duration time.Duration) {
	m.successChan <- duration
}

func (m *Metric) Fail() {
	m.failuresChan <- struct{}{}
}

func (m *Metric) Fallback() {
	m.fallbackChan <- struct{}{}
}

func (m *Metric) FallbackError() {
	m.fallbackErrorChan <- struct{}{}
}

func (m *Metric) Timeout() {
	m.timeoutsChan <- struct{}{}
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
