package goHystrix

import (
	"fmt"
	"log"
	"strings"
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

type CommandError struct {
	group         string
	name          string
	runError      error
	fallbackError error
}

// CommandOptions, you can custimize the values, for the Circuit Breaker and the Metrics stores
// errorsThreshold - if number_of_errors / total_calls * 100 > errorThreshold the circuit will be open
// minimumNumberOfRequest - if total_calls < minimumNumberOfRequest the circuit will be close
// numberOfSecondsToStore - Is the number of seconds to count the stats, for example 10 stores just the last 10 seconds of calls
// numberOfSamplesToStore - Is the number of samples to store for calculate the stats, greater means more precision to get Mean, Max, Min...
type CommandOptions struct {
	errorsThreshold        float64
	minimumNumberOfRequest int64
	numberOfSecondsToStore int
	numberOfSamplesToStore int
}

// CommandOptionsDefaults
// errorsThreshold - 50 - If number_of_errors / total_calls * 100 > 50.0 the circuit will be open
// minimumNumberOfRequest - if total_calls < 20 the circuit will be close
// numberOfSecondsToStore - 20 seconds
// numberOfSamplesToStore - 50 values
func CommandOptionsDefaults() CommandOptions {
	return CommandOptions{
		errorsThreshold:        50.0,
		minimumNumberOfRequest: 20,
		numberOfSecondsToStore: 20,
		numberOfSamplesToStore: 20,
	}

}

// NewCommand- create a new command with the default values
func NewCommand(command Interface) *Command {
	executor := NewExecutor(command)
	return &Command{command, executor}
}

func NewCommandWithOptions(command Interface, options CommandOptions) *Command {
	executor := NewExecutorWithOptions(command, options)
	return &Command{command, executor}
}

func NewExecutor(command Interface) *Executor {
	return NewExecutorWithOptions(command, CommandOptionsDefaults())
}

func NewExecutorWithOptions(command Interface, options CommandOptions) *Executor {
	circuit := NewCircuit(command.Group(), command.Name(), options)
	return &Executor{command, circuit}
}

func (ex *Executor) doExecute() (interface{}, error) {
	valueChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)
	var elapsed time.Duration
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ex.Metric().Panic()
				errorChan <- fmt.Errorf("Recovered from panic: %v", r)
			}
		}()
		start := time.Now()
		value, err := ex.command.Run()
		elapsed = time.Since(start)
		if err != nil {
			errorChan <- err
		} else {
			valueChan <- value
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

func (ex *Executor) doFallback(nestedError error) (interface{}, error) {
	ex.Metric().Fallback()

	fbCmd, ok := ex.command.(FallbackInterface)
	if !ok {
		ex.Metric().FallbackError()
		return nil, NewCommandError(ex.command.Group(), ex.command.Name(), nestedError, fmt.Errorf("No fallback implementation available for %s", ex.command.Name()))
	}

	value, err := fbCmd.Fallback()
	if err != nil {
		ex.Metric().FallbackError()
		return value, NewCommandError(ex.command.Group(), ex.command.Name(), nestedError, err)
	}

	// log the nested error
	if nestedError != nil {
		commandError := NewCommandError(ex.command.Group(), ex.command.Name(), nestedError, nil)
		log.Println(commandError.Error())
	}

	return value, err

}

func (ex *Executor) Execute() (interface{}, error) {
	open, _ := ex.circuit.IsOpen()
	if open {
		return ex.doFallback(nil)
	}

	value, err := ex.doExecute()
	if err != nil {
		return ex.doFallback(err)
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

// Nested error handling
func (e CommandError) Error() string {
	runErrorText := ""
	fallbackErrorText := ""
	commandText := fmt.Sprintf("[%s:%s]", e.group, e.name)
	if e.runError != nil {
		runErrorText = fmt.Sprintf("RunError: %s", e.runError.Error())

	}
	if e.fallbackError != nil {
		fallbackErrorText = fmt.Sprintf("FallbackError: %s", e.fallbackError.Error())
	}

	return strings.TrimSpace(fmt.Sprintf("%s %s %s", commandText, fallbackErrorText, runErrorText))
}

func NewCommandError(group string, name string, runError error, fallbackError error) CommandError {
	return CommandError{group, name, runError, fallbackError}
}
