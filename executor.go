package goHystrix

import (
	"fmt"
	"github.com/dahernan/goHystrix/sample"
	"time"
)

type Interface interface {
	Run() (interface{}, error)
	Timeout() time.Duration
	Name() string
	Group() string
}

type FallbackInterface interface {
	Interface
	Fallback() (interface{}, error)
}

type Command struct {
	Interface
	*Executor
}

type Executor struct {
	command Interface
	circuit *CircuitBreaker
}

// NewCommand- create a new command with the default values
// errorThreshold - 50 - If number_of_errors / total_calls * 100 > 50.0 the circuit will be open
// minimumNumberOfRequest - if total_calls < 20 the circuit will be close
// numberOfSecondsToStore - 20 seconds
// numberOfSamplesToStore - 50 values
func NewCommand(command Interface) *Command {
	executor := NewExecutor(command)
	return &Command{command, executor}
}

// NewCommandWithParams, you can custimize the values, for the Circuit Breaker and the Metrics stores
// errorThreshold - if number_of_errors / total_calls * 100 > errorThreshold the circuit will be open
// minimumNumberOfRequest - if total_calls < minimumNumberOfRequest the circuit will be close
// numberOfSecondsToStore - Is the number of seconds to count the stats, for example 10 stores just the last 10 seconds of calls
// numberOfSamplesToStore - Is the number of samples to store for calculate the stats, greater means more precision to get Mean, Max, Min...
func NewCommandWithParams(command Interface,
	errorThreshold float64, minimumNumberOfRequest int64, numberOfSecondsToStore int, numberOfSamplesToStore int) *Command {
	executor := NewExecutorWithParams(command, errorThreshold, minimumNumberOfRequest, numberOfSecondsToStore, numberOfSamplesToStore)
	return &Command{command, executor}
}

func NewExecutor(command Interface) *Executor {
	return NewExecutorWithParams(command, 50.0, 20, 20, 50)
}

func NewExecutorWithParams(command Interface, errorThreshold float64, minimumNumberOfRequest int64, numberOfSecondsToStore int, numberOfSamplesToStore int) *Executor {
	circuit := NewCircuit(command.Group(), command.Name(), errorThreshold, minimumNumberOfRequest, numberOfSecondsToStore, numberOfSamplesToStore)
	return &Executor{command, circuit}
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
		return nil, fmt.Errorf("error: Timeout (%s), executing command %s:%s", ex.command.Timeout(), ex.command.Group(), ex.command.Name())
	}

}

func (ex *Executor) doFallback() (interface{}, error) {
	ex.Metric().Fallback()

	if fbCmd, ok := ex.command.(FallbackInterface); ok {
		value, err := fbCmd.Fallback()
		if err != nil {
			ex.Metric().FallbackError()
		}
		return value, err
	} else {
		ex.Metric().FallbackError()
		return nil, fmt.Errorf("No fallback implementation available for %s", ex.command.Name())
	}
}

func (ex *Executor) Execute() (value interface{}, err error) {
	open, _ := ex.circuit.IsOpen()
	if !open {
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
	return ex.circuit.Metric()
}

func (ex *Executor) HealthCounts() HealthCounts {
	return ex.Metric().HealthCounts()
}

func (ex *Executor) Stats() sample.Sample {
	return ex.Metric().Stats()
}
