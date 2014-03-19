package goHystrix

import (
	"fmt"
	"github.com/dahernan/goHystrix/sample"
	"time"
)

type Interface interface {
	Run() (interface{}, error)
	Fallback() (interface{}, error)
	Timeout() time.Duration
	Name() string
	Group() string
}

type Command struct {
	Interface
	*Executor
}

type Executor struct {
	command Interface
	metric  *Metric
	circuit *CircuitBreaker
}

func NewCommand(command Interface) *Command {
	executor := NewExecutor(command)
	return &Command{command, executor}
}

func NewComandWithParams(command Interface, errorThreshold float64, minimumNumberOfRequest int64, numberOfSecondsToStore int) *Command {
	executor := NewExecutorWithParams(command, errorThreshold, minimumNumberOfRequest, numberOfSecondsToStore)
	return &Command{command, executor}
}

func NewExecutor(command Interface) *Executor {
	return NewExecutorWithParams(command, 50.0, 20, 20)
}

func NewExecutorWithParams(command Interface, errorThreshold float64, minimumNumberOfRequest int64, numberOfSecondsToStore int) *Executor {
	metric := NewMetricWithSecondsDuration(command.Group(), command.Name(), numberOfSecondsToStore)
	circuit := NewCircuit(metric, errorThreshold, minimumNumberOfRequest)
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

func (ex *Executor) Stats() sample.Sample {
	return ex.Metric().Stats()
}
