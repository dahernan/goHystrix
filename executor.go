package goHystrix

import (
	"fmt"
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
	metric  *Metric
	circuit *CircuitBreaker
}

func NewExecutor(command Command) *Executor {
	metric := NewMetric(command.Group(), command.Name())
	circuit := NewCircuit(metric, 50.0, 20)
	return &Executor{command, metric, circuit}
}

func (ex *Executor) doExecute() (interface{}, error) {
	valueChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)
	var elapsed time.Duration
	go func() {
		start := time.Now()
		value, err := ex.command.Run()
		elapsed = time.Since(start)
		if value != nil {
			valueChan <- value
		}
		if err != nil {
			errorChan <- err
		}
	}()

	select {
	case value := <-valueChan:
		ex.Metric().Success(elapsed)
		return value, nil
	case err := <-errorChan:
		ex.Metric().Fail()
		return nil, err
	case <-time.After(ex.command.Timeout()):
		ex.Metric().Timeout()
		return nil, fmt.Errorf("ERROR: Timeout!!")
	}

}

func (ex *Executor) doFallback() (interface{}, error) {
	ex.Metric().Fallback()
	value, err := ex.command.Fallback()
	if err != nil {
		ex.Metric().FallbackError()
	}
	return value, err
}

func (ex *Executor) Execute() (value interface{}, err error) {
	if !ex.circuit.IsOpen() {
		value, err = ex.doExecute()
	} else {
		value, err = ex.doFallback()
	}

	return value, err
}

func (ex *Executor) Queue() (chan interface{}, chan error) {
	valueChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)

	go func() {
		value, err := ex.Execute()
		if value != nil {
			valueChan <- value
		}
		if err != nil {
			errorChan <- err
		}
	}()
	return valueChan, errorChan
}

func (ex *Executor) Metric() *Metric {
	return ex.metric
}

func (ex *Executor) HealthCounts() HealthCounts {
	return ex.Metric().HealthCounts()
}
