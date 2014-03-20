package goHystrix

import (
	"sync"
)

type CircuitBreaker struct {
	name  string
	group string

	metric              *Metric
	errorsThreshold     float64
	minRequestThreshold int64 // min number of request
}

var (
	circuits = NewCircuitsHolder()
)

type CircuitHolder struct {
	circuits map[string]map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

func NewCircuit(group string, name string, errorsThreshold float64, minimumNumberOfRequest int64, numberOfSecondsToStore int, numberOfSamplesToStore int) *CircuitBreaker {
	c, ok := Circuits().Get(group, name)
	if ok {
		return c
	}
	metric := NewMetricWithParams(numberOfSecondsToStore, numberOfSamplesToStore)
	c = &CircuitBreaker{name: name, group: group, metric: metric, errorsThreshold: errorsThreshold, minRequestThreshold: minimumNumberOfRequest}

	Circuits().Set(group, name, c)
	return c

}

func (c *CircuitBreaker) IsOpen() (bool, string) {
	counts := c.metric.HealthCounts()

	if counts.Total < c.minRequestThreshold {
		return false, "CLOSE: not enought request"
	}

	if counts.ErrorPercentage >= c.errorsThreshold {
		return true, "OPEN: to many errors"
	}
	return false, "CLOSE: all ok"
}

func (c *CircuitBreaker) Metric() *Metric {
	return c.metric
}

func NewCircuitsHolder() *CircuitHolder {
	return &CircuitHolder{circuits: make(map[string]map[string]*CircuitBreaker)}
}

func Circuits() *CircuitHolder {
	return circuits
}
func CircuitsReset() {
	circuits = NewCircuitsHolder()
}

func (holder *CircuitHolder) Get(group string, name string) (*CircuitBreaker, bool) {
	holder.mutex.RLock()
	defer holder.mutex.RUnlock()
	circuitsValues, ok := holder.circuits[group]
	if !ok {
		return nil, ok
	}

	value, ok := circuitsValues[name]
	return value, ok
}

func (holder *CircuitHolder) Set(group string, name string, value *CircuitBreaker) {
	holder.mutex.Lock()
	defer holder.mutex.Unlock()

	circuitsValues, ok := holder.circuits[group]
	if !ok {
		circuitsValues = make(map[string]*CircuitBreaker)
		holder.circuits[group] = circuitsValues
	}
	circuitsValues[name] = value
}
