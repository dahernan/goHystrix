package goHystrix

type CircuitBreaker struct {
	metric              *Metric
	errorsThreshold     float64
	minRequestThreshold int64 // min number of request
}

func NewCircuit(metric *Metric, errorsThreshold float64, min int64) *CircuitBreaker {
	return &CircuitBreaker{metric: metric, errorsThreshold: errorsThreshold, minRequestThreshold: min}
}

func (c *CircuitBreaker) IsOpen() bool {
	counts := c.metric.HealthCounts()

	if counts.Total < c.minRequestThreshold {
		return false
	}

	if counts.ErrorPercentage >= c.errorsThreshold {
		return true
	}
	return false
}
