package goHystrix

import (
	"github.com/dahernan/goHystrix/sample"
	//"github.com/dahernan/goHystrix/statsd"

	"time"
)

const (
	alpha = 0.015 // alpha for the exponential decay distribution
)

type Metric struct {
	name  string
	group string

	successChan       chan time.Duration
	failuresChan      chan struct{}
	fallbackChan      chan struct{}
	fallbackErrorChan chan struct{}
	timeoutsChan      chan struct{}
	panicChan         chan struct{}
	countersChan      chan struct{}
	countersOutChan   chan HealthCounts

	buckets int
	window  time.Duration
	values  []HealthCountsBucket

	sample sample.Sample

	lastFailure time.Time
	lastSuccess time.Time
	lastTimeout time.Time
}

func NewMetric(group string, name string) *Metric {
	return NewMetricWithParams(group, name, 20, 50)
}

func NewMetricWithParams(group string, name string, numberOfSecondsToStore int, sampleSize int) *Metric {
	m := &Metric{}
	m.name = name
	m.group = group
	m.buckets = numberOfSecondsToStore
	m.window = time.Duration(numberOfSecondsToStore) * time.Second
	m.values = make([]HealthCountsBucket, m.buckets, m.buckets)

	m.sample = sample.NewExpDecaySample(sampleSize, alpha)

	m.successChan = make(chan time.Duration)
	m.failuresChan = make(chan struct{})
	m.fallbackChan = make(chan struct{})
	m.fallbackErrorChan = make(chan struct{})
	m.timeoutsChan = make(chan struct{})
	m.panicChan = make(chan struct{})
	m.countersChan = make(chan struct{})
	m.countersOutChan = make(chan HealthCounts)

	go m.run()
	return m

}

type HealthCountsBucket struct {
	Failures       int64
	Success        int64
	Fallback       int64
	FallbackErrors int64
	Timeouts       int64
	Panics         int64
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
	c.Panics = 0
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
		case <-m.panicChan:
			m.doPanic()
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

	go func(d time.Duration) {
		m.sample.Update(int64(d))
	}(duration)
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

func (m *Metric) doPanic() {
	m.bucket().Panics++
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
			counters.Panics += value.Panics
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

func (m *Metric) Panic() {
	m.panicChan <- struct{}{}
}

func (m *Metric) Stats() sample.Sample {
	return m.sample
}

func (m *Metric) LastFailure() time.Time {
	return m.lastFailure
}
func (m *Metric) LastSuccess() time.Time {
	return m.lastSuccess
}
func (m *Metric) LastTimeout() time.Time {
	return m.lastTimeout
}
