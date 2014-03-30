package goHystrix

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

type StringCommand struct {
	state         string
	fallbackState string
}

func NewStringCommand(state string, fallbackState string) *Command {
	var command *StringCommand
	command = &StringCommand{}
	command.state = state
	command.fallbackState = fallbackState

	return NewCommandWithParams(command, 50.0, 3, 5, 10)
}

func (c *StringCommand) Name() string {
	return "testCommand"
}

func (c *StringCommand) Group() string {
	return "testGroup"
}

func (c *StringCommand) Timeout() time.Duration {
	return 3 * time.Millisecond
}

func (c *StringCommand) Run() (interface{}, error) {
	if c.state == "error" {
		return nil, fmt.Errorf("ERROR: this method is mend to fail")
	}

	if c.state == "timeout" {
		time.Sleep(4 * time.Millisecond)
		return "time out!", nil
	}

	return "hello hystrix world", nil
}

func (c *StringCommand) Fallback() (interface{}, error) {
	if c.fallbackState == "fallbackError" {
		return nil, fmt.Errorf("ERROR: error doing fallback")
	}
	return "FALLBACK", nil

}

func TestRunString(t *testing.T) {

	Convey("Command Run returns a string", t, func() {
		CircuitsReset()
		okCommand := NewStringCommand("ok", "fallbackOk")

		Convey("When Run is executed", func() {

			result, err := okCommand.Run()

			Convey("The result should be the string value", func() {
				So(result, ShouldEqual, "hello hystrix world")
			})

			Convey("There is no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestRunError(t *testing.T) {

	Convey("Command Run returns an error", t, func() {
		CircuitsReset()
		errorCommand := NewStringCommand("error", "fallbackOk")

		Convey("When Run is executed", func() {
			result, err := errorCommand.Run()

			Convey("The result should be Nil", func() {
				So(result, ShouldBeNil)
			})

			Convey("There is a expected error", func() {
				So(err.Error(), ShouldEqual, "ERROR: this method is mend to fail")
			})
		})
	})
}

func TestExecuteString(t *testing.T) {

	Convey("Command Execute runs properly", t, func() {
		CircuitsReset()
		okCommand := NewStringCommand("ok", "fallbackOk")

		Convey("When Execute is called", func() {

			result, err := okCommand.Execute()

			Convey("The result should be the string value", func() {
				So(result, ShouldEqual, "hello hystrix world")
			})

			Convey("There is no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestFallbackForError(t *testing.T) {
	Convey("Command Execute uses the Fallback if an error is returned", t, func() {
		CircuitsReset()
		errorCommand := NewStringCommand("error", "fallbackOk")

		var result interface{}
		var err error

		// 1
		result, err = errorCommand.Execute()
		So(err, ShouldBeNil)
		So(result, ShouldEqual, "FALLBACK")
		open, reason := errorCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "CLOSE: not enought request")
		So(open, ShouldBeFalse)
		So(errorCommand.HealthCounts().Failures, ShouldEqual, 1)
		So(errorCommand.HealthCounts().Fallback, ShouldEqual, 1)
	})
}

func TestFallbackForTimeout(t *testing.T) {
	Convey("Command Execute uses the Fallback if a timeout is returned", t, func() {
		CircuitsReset()
		timeoutCommand := NewStringCommand("timeout", "fallbackOk")

		var result interface{}
		var err error

		// 1
		result, err = timeoutCommand.Execute()
		So(err, ShouldBeNil)
		So(result, ShouldEqual, "FALLBACK")
		open, reason := timeoutCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "CLOSE: not enought request")
		So(open, ShouldBeFalse)
		So(timeoutCommand.HealthCounts().Timeouts, ShouldEqual, 1)
		So(timeoutCommand.HealthCounts().Fallback, ShouldEqual, 1)
	})
}

func TestFallback(t *testing.T) {

	Convey("Command Execute failing for 3 times, opens the circuit", t, func() {
		CircuitsReset()
		errorCommand := NewStringCommand("error", "fallbackOk")

		// 1
		errorCommand.Execute()
		open, reason := errorCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "CLOSE: not enought request")
		So(open, ShouldBeFalse)

		// 2
		errorCommand.Execute()
		open, reason = errorCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "CLOSE: not enought request")
		So(open, ShouldBeFalse)

		//3
		errorCommand.Execute()
		// limit reached
		open, reason = errorCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "OPEN: to many errors")
		So(open, ShouldBeTrue)
		So(errorCommand.HealthCounts().Failures, ShouldEqual, 3)

	})
}

func TestExecuteTimeout(t *testing.T) {

	Convey("Command returns the fallback due to timeout", t, func() {
		CircuitsReset()
		timeoutCommand := NewStringCommand("timeout", "fallbackOk")

		// 1
		timeoutCommand.Execute()
		open, reason := timeoutCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "CLOSE: not enought request")
		So(open, ShouldBeFalse)

		// 2
		timeoutCommand.Execute()
		open, reason = timeoutCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "CLOSE: not enought request")
		So(open, ShouldBeFalse)

		//3
		timeoutCommand.Execute()
		// limit reached
		open, reason = timeoutCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "OPEN: to many errors")
		So(open, ShouldBeTrue)
		So(timeoutCommand.HealthCounts().Failures, ShouldEqual, 3)
		So(timeoutCommand.HealthCounts().Timeouts, ShouldEqual, 3)

	})

}
func TestAsync(t *testing.T) {
	Convey("Command run async and returns ok", t, func() {
		CircuitsReset()
		okCommand := NewStringCommand("ok", "fallbackOk")

		resultChan, errorChan := okCommand.Queue()
		var err error
		var result interface{}
		select {
		case result = <-resultChan:
			err = nil
		case err = <-errorChan:
			result = nil
		}

		So(result, ShouldEqual, "hello hystrix world")
		So(err, ShouldBeNil)

	})
}

func TestAsyncFallback(t *testing.T) {

	Convey("Command run async, and if it's failing 3 times, the Circuit will be open", t, func() {
		CircuitsReset()
		errorCommand := NewStringCommand("error", "fallbackOk")

		var result interface{}

		// 1 fail
		resultChan, _ := errorCommand.Queue()
		result = <-resultChan
		So(result, ShouldEqual, "FALLBACK")
		open, reason := errorCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "CLOSE: not enought request")
		So(open, ShouldBeFalse)

		// 2  fail
		resultChan, _ = errorCommand.Queue()
		result = <-resultChan
		So(result, ShouldEqual, "FALLBACK")
		open, reason = errorCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "CLOSE: not enought request")
		So(open, ShouldBeFalse)

		// 3 fail
		resultChan, _ = errorCommand.Queue()
		result = <-resultChan
		So(result, ShouldEqual, "FALLBACK")
		So(errorCommand.HealthCounts().Failures, ShouldEqual, 3)

		// limit reached
		open, reason = errorCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "OPEN: to many errors")
		So(open, ShouldBeTrue)

		// 4 falling back
		resultChan, _ = errorCommand.Queue()
		result = <-resultChan
		So(result, ShouldEqual, "FALLBACK")
		So(errorCommand.HealthCounts().Failures, ShouldEqual, 3)

	})
}

func TestAsyncTimeout(t *testing.T) {
	Convey("Command run async and if it is timeouting for 3 times the Circuit will be open", t, func() {
		var result interface{}

		CircuitsReset()
		timeoutCommand := NewStringCommand("timeout", "fallbackOk")

		// 1 timeout
		resultChan, _ := timeoutCommand.Queue()
		result = <-resultChan
		So(result, ShouldEqual, "FALLBACK")
		open, reason := timeoutCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "CLOSE: not enought request")
		So(open, ShouldBeFalse)

		// 2 timeout
		resultChan, _ = timeoutCommand.Queue()
		result = <-resultChan

		So(result, ShouldEqual, "FALLBACK")
		open, reason = timeoutCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "CLOSE: not enought request")
		So(open, ShouldBeFalse)

		// 3 timeout
		resultChan, _ = timeoutCommand.Queue()
		result = <-resultChan
		So(result, ShouldEqual, "FALLBACK")
		open, reason = timeoutCommand.circuit.IsOpen()
		So(reason, ShouldEqual, "OPEN: to many errors")
		So(open, ShouldBeTrue)

		// 4 falling back
		resultChan, _ = timeoutCommand.Queue()
		result = <-resultChan

		So(result, ShouldEqual, "FALLBACK")
		So(timeoutCommand.HealthCounts().Failures, ShouldEqual, 3)
		So(timeoutCommand.HealthCounts().Timeouts, ShouldEqual, 3)
	})

}

func TestMetrics(t *testing.T) {
	Convey("Command keep the metrics", t, func() {
		CircuitsReset()
		x := NewStringCommand("ok", "fallbackok")
		y := NewStringCommand("error", "fallbackok")

		Convey("When Execute is called multiple times the counters are updated", func() {
			x.Execute() // success
			x.Execute() // success
			y.Execute() // error
			y.Execute() // error
			y.Execute() // fallback

			Convey("The success and failures counters are correct", func() {
				So(x.HealthCounts().Success, ShouldEqual, 2)
				So(y.HealthCounts().Success, ShouldEqual, 2)
				So(x.HealthCounts().Failures, ShouldEqual, 2)
				So(y.HealthCounts().Failures, ShouldEqual, 2)
				So(x.HealthCounts().Fallback, ShouldEqual, 3)
				So(y.HealthCounts().Fallback, ShouldEqual, 3)

				fmt.Println(x.circuit.ToJSON())

			})

		})

	})
}
