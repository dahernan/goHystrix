package goHystrix

import (
	"fmt"
	"github.com/dahernan/goHystrix/metrics"
	"time"
)

type Command interface {
	Run() (interface{}, error)
	Fallback() (interface{}, error)
	Timeout() time.Duration
	Name() string
	Group() string
}

type Executor struct {
	command Command
	metric  *metrics.Metric
}

func NewExecutor(command Command) *Executor {
	metric, ok := metrics.Metrics().Get(command.Group(), command.Name())
	if !ok {
		metric = metrics.NewMetric(command.Group(), command.Name())
	}
	return &Executor{command, metric}
}

func (h *Executor) doExecute() (interface{}, error) {
	valueChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)
	go func() {
		value, err := h.command.Run()
		if value != nil {
			valueChan <- value
		}
		if err != nil {
			h.Metric().Fail()
			errorChan <- err
		} else {
			h.Metric().Success()
		}
	}()

	select {
	case value := <-valueChan:
		return value, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(h.command.Timeout()):
		h.Metric().Timeout()
		return nil, fmt.Errorf("ERROR: Timeout!!")
	}

}

func (h *Executor) doFallback() (interface{}, error) {
	h.Metric().Fallback()
	value, err := h.command.Fallback()
	if err != nil {
		h.Metric().FallbackError()
	}
	return value, err
}

func (h *Executor) Execute() (interface{}, error) {
	start := time.Now()
	value, err := h.doExecute()
	if err != nil {
		return h.doFallback()
	}
	elapsed := time.Since(start)
	fmt.Printf("It took %s\n", elapsed)
	return value, err
}

func (h *Executor) Queue() (chan interface{}, chan error) {
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

func (h *Executor) Metric() *metrics.Metric {
	return h.metric
}

func (h *Executor) SuccessCount() int64 {
	return h.Metric().SuccessCount()
}

func (h *Executor) FailuresCount() int64 {
	return h.Metric().FailuresCount()
}
