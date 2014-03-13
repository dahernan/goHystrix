package goHystrix

import (
	"github.com/dahernan/goHystrix/metrics"
)

//type CircuitBreaker struct {
//}

type Circuit interface {
	IsOpen() bool
}

type CircuitBreaker struct {
	metric *metrics.Metric
	state  string
}

func NewCircuit(metric *metrics.Metric) Circuit {
	return &CircuitBreaker{metric: metric, state: "close"}
}

func (c *CircuitBreaker) IsOpen() bool {
	return false
}
