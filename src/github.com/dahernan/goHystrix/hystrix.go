package goHystrix

import (
	"fmt"
	"github.com/dahernan/goHystrix/metrics"
	"time"
)

type HystrixCommand interface {
	Run() (interface{}, error)
	Fallback() (interface{}, error)
	Name() string
	Group() string
}

type HystrixExecutor struct {
	command HystrixCommand
}

func NewHystrixExecutor(command HystrixCommand) *HystrixExecutor {
	return &HystrixExecutor{command}
}

func (h *HystrixExecutor) doExecute() (interface{}, error) {
	valueChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)
	go func() {
		value, err := h.command.Run()
		if value != nil {
			h.metric().Success.Inc(1)
			valueChan <- value
		}
		if err != nil {
			h.metric().Failures.Inc(1)
			errorChan <- err
		}
	}()

	select {
	case value := <-valueChan:
		return value, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(2 * time.Second):
		return nil, fmt.Errorf("ERROR: Timeout!!")
	}

}
func (h *HystrixExecutor) metric() *metrics.Metric {
	metric, ok := metrics.Metrics().Get(h.command.Group(), h.command.Name())
	if !ok {
		return metrics.NewMetric(h.command.Name(), h.command.Group())
	}
	return metric
}

func (h *HystrixExecutor) Execute() (interface{}, error) {
	start := time.Now()
	value, err := h.doExecute()
	if err != nil {
		return h.command.Fallback()
	}
	elapsed := time.Since(start)
	fmt.Printf("It took %s\n", elapsed)
	return value, err
}

func (h *HystrixExecutor) Queue() (chan interface{}, chan error) {
	valueChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)

	go func() {
		value, err := h.Execute()
		if value != nil {
			valueChan <- value
		}
		if err != nil {
			errorChan <- err
		}
	}()
	return valueChan, errorChan
}

func (h *HystrixExecutor) Success() int64 {
	return h.metric().Success.Count()
}

func (h *HystrixExecutor) Failures() int64 {
	return h.metric().Failures.Count()
}
