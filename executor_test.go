package goHystrix

import (
	"fmt"
	"github.com/dahernan/goHystrix/metrics"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

type StringCommand struct {
	state         string
	fallbackState string
	*Executor
}

func NewStringCommand(state string, fallbackState string) *StringCommand {
	var command *StringCommand
	command = &StringCommand{}
	executor := NewExecutor(command)
	command.Executor = executor
	command.state = state
	command.fallbackState = fallbackState
	return command
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
		x := NewStringCommand("ok", "fallbackOk")

		Convey("When Run is executed", func() {

			result, err := x.Run()

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
		x := NewStringCommand("error", "fallbackOk")

		Convey("When Run is executed", func() {
			result, err := x.Run()

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
		x := NewStringCommand("ok", "fallbackOk")

		Convey("When Execute is called", func() {

			result, err := x.Execute()

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
		x := NewStringCommand("error", "fallbackOk")

		Convey("When Execute is called", func() {
			result, err := x.Execute()

			Convey("The result should be the one from the fallback function", func() {
				So(result, ShouldEqual, "FALLBACK")
			})

			Convey("There is no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestExecuteTimeout(t *testing.T) {

	Convey("Command returns the fallback due to timeout", t, func() {
		x := NewStringCommand("timeout", "fallbackOk")

		Convey("When Execute is called", func() {
			result, err := x.Execute()

			Convey("The result should be FALLBACK", func() {
				So(result, ShouldEqual, "FALLBACK")
			})

			Convey("There is no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})

}
func TestAsync(t *testing.T) {
	Convey("Command run async and returns ok", t, func() {
		x := NewStringCommand("ok", "fallbackOk")

		Convey("When Queue is called", func() {
			resultChan, errorChan := x.Queue()
			var err error
			var result interface{}
			select {
			case result = <-resultChan:
				err = nil
			case err = <-errorChan:
				result = nil
			}

			Convey("The result should be the string value", func() {
				So(result, ShouldEqual, "hello hystrix world")
			})
			Convey("There is no error", func() {
				So(err, ShouldBeNil)
			})

		})
	})
}

func TestAsyncFallback(t *testing.T) {

	Convey("Command run async and returns the fallback", t, func() {
		x := NewStringCommand("error", "fallbackOk")

		Convey("When Queue is called", func() {
			resultChan, errorChan := x.Queue()
			var err error
			var result interface{}
			select {
			case result = <-resultChan:
				err = nil
			case err = <-errorChan:
				result = nil
			}

			Convey("The result should be the fallback", func() {
				So(result, ShouldEqual, "FALLBACK")
			})
			Convey("There is no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestAsyncTimeout(t *testing.T) {
	Convey("Command run async and returns the fallback due a timeout error", t, func() {
		metrics.MetricsReset()

		x := NewStringCommand("timeout", "fallbackOk")

		Convey("When Queue is executed", func() {
			resultChan, errorChan := x.Queue()
			var err error
			var result interface{}
			select {
			case result = <-resultChan:
				err = nil
			case err = <-errorChan:
				result = nil
			}

			Convey("The result should be the fallback string value and there is not error", func() {
				So(result, ShouldEqual, "FALLBACK")
				So(x.Metric().TimeoutsCount(), ShouldEqual, 1)
				So(err, ShouldBeNil)

			})
		})
	})

}

func TestAsyncFallbackError(t *testing.T) {

	Convey("Command run async and returns the fallback error", t, func() {
		x := NewStringCommand("error", "fallbackError")

		Convey("When Queue is executed", func() {
			resultChan, errorChan := x.Queue()
			var err error
			var result interface{}
			select {
			case result = <-resultChan:
				err = nil
			case err = <-errorChan:
				result = nil
			}

			Convey("The result should be the nil", func() {
				So(result, ShouldBeNil)
			})
			Convey("There is an error from the fallback", func() {
				So(err.Error(), ShouldEqual, "ERROR: error doing fallback")
			})
		})
	})

}

func TestMetrics(t *testing.T) {
	Convey("Command keep the metrics", t, func() {
		metrics.MetricsReset()
		x := NewStringCommand("ok", "fallbackok")
		y := NewStringCommand("error", "fallbackok")

		Convey("When Execute is called 2 times the counters are updated", func() {
			x.Execute()
			x.Execute()
			y.Execute()
			y.Execute()
			y.Execute()

			Convey("The success and failures counters are correct", func() {
				So(x.SuccessCount(), ShouldEqual, 2)
				So(y.SuccessCount(), ShouldEqual, 2)
				So(x.FailuresCount(), ShouldEqual, 3)
				So(y.FailuresCount(), ShouldEqual, 3)
			})

		})

	})
}
