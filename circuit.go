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
	metric          *metrics.Metric
	errorsThreshold float64
}

func NewCircuit(metric *metrics.Metric, errorsThreshold float64) Circuit {
	return &CircuitBreaker{metric: metric, errorsThreshold: errorsThreshold}
}

func (c *CircuitBreaker) IsOpen() bool {
	counts := c.metric.HealthCounts()

	fmt.Println("ErrorPercentage  failures: ", counts)
	if counts.ErrorPercentage >= c.errorsThreshold {
		return true
	}

	return false
}
