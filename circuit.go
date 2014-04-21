package goHystrix

import (
	"bytes"
	"fmt"
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

func NewCircuitNoParams(group string, name string) *CircuitBreaker {
	return NewCircuit(group, name, 50.0, 20, 20, 50)
}

func NewCircuit(group string, name string, errorsThreshold float64, minimumNumberOfRequest int64, numberOfSecondsToStore int, numberOfSamplesToStore int) *CircuitBreaker {
	c, ok := Circuits().Get(group, name)
	if ok {
		return c
	}
	metric := NewMetricWithParams(group, name, numberOfSecondsToStore, numberOfSamplesToStore)
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

func (c *CircuitBreaker) ToJSON() string {

	var buffer bytes.Buffer

	buffer.WriteString("{\n")

	open, state := c.IsOpen()
	counts := c.Metric().doHealthCounts()
	stats := c.Metric().Stats()
	lastSuccess := c.Metric().LastSuccess()
	lastFailure := c.Metric().LastFailure()
	lastTimeout := c.Metric().LastTimeout()

	fmt.Fprintf(&buffer, "\"name\" : \"%s\",\n", c.name)
	fmt.Fprintf(&buffer, "\"group\" : \"%s\",\n", c.group)

	fmt.Fprintf(&buffer, "\"isOpen\" : \"%t\",\n", open)
	fmt.Fprintf(&buffer, "\"state\" : \"%s\",\n", state)

	fmt.Fprintf(&buffer, "\"percentile90\" : \"%f\",\n", stats.Percentile(0.90))
	fmt.Fprintf(&buffer, "\"mean\" : \"%f\",\n", stats.Mean())
	fmt.Fprintf(&buffer, "\"variance\" : \"%f\",\n", stats.Variance())

	fmt.Fprintf(&buffer, "\"max\" : \"%d\",\n", stats.Max())
	fmt.Fprintf(&buffer, "\"min\" : \"%d\",\n", stats.Min())

	fmt.Fprintf(&buffer, "\"failures\" : \"%d\",\n", counts.Failures)
	fmt.Fprintf(&buffer, "\"timeouts\" : \"%d\",\n", counts.Timeouts)
	fmt.Fprintf(&buffer, "\"fallback\" : \"%d\",\n", counts.Fallback)
	fmt.Fprintf(&buffer, "\"panics\" : \"%d\",\n", counts.Panics)
	fmt.Fprintf(&buffer, "\"fallbackErrors\" : \"%d\",\n", counts.FallbackErrors)
	fmt.Fprintf(&buffer, "\"total\" : \"%d\",\n", counts.Total)
	fmt.Fprintf(&buffer, "\"success\" : \"%d\",\n", counts.Success)
	fmt.Fprintf(&buffer, "\"errorPercentage\" : \"%f\",\n", counts.ErrorPercentage)

	fmt.Fprintf(&buffer, "\"lastSuccess\" : \"%s\",\n", lastSuccess)
	fmt.Fprintf(&buffer, "\"lastFailure\" : \"%s\",\n", lastFailure)
	fmt.Fprintf(&buffer, "\"lastTimeout\" : \"%s\"\n", lastTimeout)

	buffer.WriteString("\n}")

	return buffer.String()

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

func (holder *CircuitHolder) ToJSON() string {
	holder.mutex.RLock()
	defer holder.mutex.RUnlock()

	var buffer bytes.Buffer

	buffer.WriteString("[\n")

	first := true
	for group, names := range holder.circuits {
		if !first {
			fmt.Fprintf(&buffer, ",\n")
		}
		first = false
		nested_first := true
		fmt.Fprintf(&buffer, "{\"group\" : \"%s\",\n", group)
		fmt.Fprintf(&buffer, "\"circuit\" : [\n")
		for _, circuit := range names {
			if !nested_first {
				fmt.Fprintf(&buffer, ",\n")
			}
			nested_first = false
			buffer.WriteString(circuit.ToJSON())
		}
		fmt.Fprintf(&buffer, "] }\n")
	}

	buffer.WriteString("\n]")
	return buffer.String()
}
