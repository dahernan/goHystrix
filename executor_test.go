package goHystrix

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

type NoFallbackCommand struct {
	state string
}

func (cmd *NoFallbackCommand) Name() string           { return "nofallbackCmd" }
func (cmd *NoFallbackCommand) Group() string          { return "testGroup" }
func (cmd *NoFallbackCommand) Timeout() time.Duration { return 3 * time.Second }
func (cmd *NoFallbackCommand) Run() (interface{}, error) {
	return "", fmt.Errorf(cmd.state)
}

func TestRunNoFallback(t *testing.T) {
	Convey("Command Execute errors directly, without fallback implementation", t, func() {
		CircuitsReset()
		errorCommand := NewCommandWithParams(&NoFallbackCommand{"error"}, 50.0, 3, 5, 10)

		Convey("After 3 errors, the circuit is open and the next call is using the fallback", func() {
			var result interface{}
			var err error

			// 1
			result, err = errorCommand.Execute()
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)

			// 2
			result, err = errorCommand.Execute()
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)

			//3
			result, err = errorCommand.Execute()
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)

			// 4 limit reached, falling back
			result, err = errorCommand.Execute()
			So(err.Error(), ShouldEqual, "No fallback implementation available for nofallbackCmd")
			So(result, ShouldBeNil)
			So(errorCommand.HealthCounts().Failures, ShouldEqual, 3)

		})
	})
}

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

func TestFallback(t *testing.T) {

	Convey("Command Execute uses the Fallback", t, func() {
		CircuitsReset()
		errorCommand := NewStringCommand("error", "fallbackOk")

		Convey("After 3 errors, the circuit is open and the next call is using the fallback", func() {
			var result interface{}
			var err error

			// 1
			result, err = errorCommand.Execute()
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)

			// 2
			result, err = errorCommand.Execute()
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)

			//3
			result, err = errorCommand.Execute()
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)

			// 4 limit reached, falling back
			result, err = errorCommand.Execute()
			So(err, ShouldBeNil)
			So(result, ShouldEqual, "FALLBACK")
			So(errorCommand.HealthCounts().Failures, ShouldEqual, 3)

		})
	})
}

func TestExecuteTimeout(t *testing.T) {

	Convey("Command returns the fallback due to timeout", t, func() {
		CircuitsReset()
		timeoutCommand := NewStringCommand("timeout", "fallbackOk")

		var result interface{}
		var err error

		// 1
		result, err = timeoutCommand.Execute()
		So(err, ShouldNotBeNil)
		//So(err, ShouldEqual, "error: Timeout, executing command testGroup:testCommand")
		So(result, ShouldBeNil)

		// 2
		result, err = timeoutCommand.Execute()
		So(err, ShouldNotBeNil)
		So(result, ShouldBeNil)

		//3
		result, err = timeoutCommand.Execute()
		So(err, ShouldNotBeNil)
		So(result, ShouldBeNil)

		// 4 limit reached, falling back
		result, err = timeoutCommand.Execute()
		So(err, ShouldBeNil)
		So(result, ShouldEqual, "FALLBACK")
		So(timeoutCommand.HealthCounts().Failures, ShouldEqual, 3)
		So(timeoutCommand.HealthCounts().Timeouts, ShouldEqual, 3)

	})

}
func TestAsync(t *testing.T) {
	Convey("Command run async and returns ok", t, func() {
		CircuitsReset()
		okCommand := NewStringCommand("ok", "fallbackOk")

		Convey("When Queue is called the result should be ok", func() {
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
	})
}

func TestAsyncFallback(t *testing.T) {

	Convey("Command run async and returns the fallback", t, func() {
		CircuitsReset()
		errorCommand := NewStringCommand("error", "fallbackOk")

		Convey("When Queue is called 3 times, the next time runs the fallback", func() {
			var err error
			var result interface{}

			// 1 fail
			resultChan, errorChan := errorCommand.Queue()
			err = <-errorChan
			So(err, ShouldNotBeNil)

			// 2  fail
			resultChan, errorChan = errorCommand.Queue()
			err = <-errorChan
			So(err, ShouldNotBeNil)

			// 3 fail
			resultChan, errorChan = errorCommand.Queue()
			err = <-errorChan
			So(err, ShouldNotBeNil)

			// 4 falling back
			resultChan, errorChan = errorCommand.Queue()

			select {
			case result = <-resultChan:
				err = nil
			case err = <-errorChan:
				result = nil
			}

			So(err, ShouldBeNil)
			So(result, ShouldEqual, "FALLBACK")
			So(errorCommand.HealthCounts().Failures, ShouldEqual, 3)

		})
	})
}

func TestAsyncTimeout(t *testing.T) {
	Convey("Command run async and returns the fallback due a timeout error", t, func() {
		var err error
		var result interface{}

		CircuitsReset()
		timeoutCommand := NewStringCommand("timeout", "fallbackOk")

		// 1 timeout
		resultChan, errorChan := timeoutCommand.Queue()
		err = <-errorChan
		So(err, ShouldNotBeNil)

		// 2 timeout
		resultChan, errorChan = timeoutCommand.Queue()
		err = <-errorChan
		So(err, ShouldNotBeNil)

		// 3 timeout
		resultChan, errorChan = timeoutCommand.Queue()
		err = <-errorChan
		So(err, ShouldNotBeNil)

		// 4 falling back
		resultChan, errorChan = timeoutCommand.Queue()

		select {
		case result = <-resultChan:
			err = nil
		case err = <-errorChan:
			result = nil
		}

		So(err, ShouldBeNil)
		So(result, ShouldEqual, "FALLBACK")
		So(timeoutCommand.HealthCounts().Failures, ShouldEqual, 3)
		So(timeoutCommand.HealthCounts().Timeouts, ShouldEqual, 3)
	})

}

func TestAsyncFallbackError(t *testing.T) {

	Convey("Command run async and returns the fallback error after 3 times falling", t, func() {
		CircuitsReset()
		fallbackErrorCommand := NewStringCommand("error", "fallbackError")

		var err error
		var result interface{}

		// 1 fail
		resultChan, errorChan := fallbackErrorCommand.Queue()
		err = <-errorChan
		So(err, ShouldNotBeNil)

		// 2 fail
		resultChan, errorChan = fallbackErrorCommand.Queue()
		err = <-errorChan
		So(err, ShouldNotBeNil)

		// 3 fail
		resultChan, errorChan = fallbackErrorCommand.Queue()
		err = <-errorChan
		So(err, ShouldNotBeNil)

		// 4 falling back error
		resultChan, errorChan = fallbackErrorCommand.Queue()

		select {
		case result = <-resultChan:
			err = nil
		case err = <-errorChan:
			result = nil
		}

		So(err.Error(), ShouldEqual, "ERROR: error doing fallback")
		So(result, ShouldBeNil)
		So(fallbackErrorCommand.HealthCounts().Failures, ShouldEqual, 3)
		So(fallbackErrorCommand.HealthCounts().FallbackErrors, ShouldEqual, 1)
		So(fallbackErrorCommand.HealthCounts().Timeouts, ShouldEqual, 0)

	})

}

func TestMetrics(t *testing.T) {
	Convey("Command keep the metrics", t, func() {
		CircuitsReset()
		x := NewStringCommand("ok", "fallbackok")
		y := NewStringCommand("error", "fallbackok")

		Convey("When Execute is called 2 times the counters are updated", func() {
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
				So(x.HealthCounts().Fallback, ShouldEqual, 1)
				So(y.HealthCounts().Fallback, ShouldEqual, 1)

				fmt.Println(x.circuit.ToJSON())

			})

		})

	})
}
