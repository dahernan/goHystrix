package goHystrix

import (
	"fmt"
	"time"
)

type HystrixCommand interface {
	Run() (interface{}, error)
	Fallback() (interface{}, error)
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
			valueChan <- value
		}
		if err != nil {
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

func (h *HystrixExecutor) Execute() (interface{}, error) {
	value, err := h.doExecute()
	if err != nil {
		return h.command.Fallback()
	}
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
