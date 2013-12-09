package goHystrix

import (
	"fmt"
	"time"
)

type HystrixCommand interface {
	Run() (interface{}, error)
}

type HystrixExecutor struct {
	command HystrixCommand
}

func NewHystrixExecutor(command HystrixCommand) *HystrixExecutor {
	return &HystrixExecutor{command}
}

func (h *HystrixExecutor) Execute() (interface{}, error) {
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
		fmt.Printf("Recieve a value '%s'\n", value)
		return value, nil
	case err := <-errorChan:
		fmt.Printf("Recieve a error '%s'\n", err)
		return nil, err
	case <-time.After(2 * time.Second):
		fmt.Printf("Recieve a timeout\n")
		return nil, fmt.Errorf("ERROR: Timeout!!")
	}

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
