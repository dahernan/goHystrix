package goHystrix

import (
	"fmt"
	"github.com/dahernan/goHystrix/metrics"
)

//type CircuitBreaker struct {
//}

type Circuit interface {
	IsOpen() bool
}

type CircuitBreaker struct {
	metric            *metrics.Metric
	state             string
	failuresThreshold int64
}

func NewCircuit(metric *metrics.Metric, failuresThreshold int64) Circuit {
	return &CircuitBreaker{metric: metric, state: "close", failuresThreshold: failuresThreshold}
}

func (c *CircuitBreaker) IsOpen() bool {
	fmt.Println("Consecutive failures: ", c.metric.ConsecutiveFailures())
	if c.metric.ConsecutiveFailures() >= c.failuresThreshold {
		return true
	}

	return false
}
