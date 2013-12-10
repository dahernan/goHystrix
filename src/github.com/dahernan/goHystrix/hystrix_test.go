package goHystrix

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

type HystrixStringCommand struct {
	state         string
	fallbackState string
	*HystrixExecutor
}

func NewHystrixStringCommand(state string, fallbackState string) *HystrixStringCommand {
	command := &HystrixStringCommand{}
	executor := NewHystrixExecutor(command)
	command.HystrixExecutor = executor
	command.state = state
	command.fallbackState = fallbackState
	return command
}

func (h *HystrixStringCommand) Run() (interface{}, error) {
	if h.state == "error" {
		return nil, fmt.Errorf("ERROR: this method is mend to fail")
	}

	if h.state == "timeout" {
		time.Sleep(3 * time.Second)
		return "time out!", nil
	}

	return "hello hystrix world", nil
}

func (h *HystrixStringCommand) Fallback() (interface{}, error) {
	if h.fallbackState == "fallbackError" {
		return nil, fmt.Errorf("ERROR: error doing fallback")
	}
	return "FALLBACK", nil

}

func TestRunsOk(t *testing.T) {

	Convey("Hytrix command Run returns a string", t, func() {
		x := NewHystrixStringCommand("ok", "fallbackOk")

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

	Convey("Hytrix command Run returns an error", t, func() {
		x := NewHystrixStringCommand("error", "fallbackOk")

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

	Convey("Hytrix command Execute runs properly", t, func() {
		x := NewHystrixStringCommand("ok", "fallbackOk")

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

	Convey("Hytrix command Execute uses the Fallback", t, func() {
		x := NewHystrixStringCommand("error", "fallbackOk")

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

	Convey("Hytrix command returns the fallback due to timeout", t, func() {
		x := NewHystrixStringCommand("timeout", "fallbackOk")

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

	Convey("Hytrix command run async and returns ok", t, func() {
		x := NewHystrixStringCommand("ok", "fallbackOk")

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

	Convey("Hytrix command run async and returns the fallback", t, func() {
		x := NewHystrixStringCommand("error", "fallbackOk")

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

	Convey("Hytrix command run async and returns the fallback due a timeout error", t, func() {
		x := NewHystrixStringCommand("timeout", "fallbackOk")

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

			Convey("The result should be the fallback string value", func() {
				So(result, ShouldEqual, "FALLBACK")
			})
			Convey("There is no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Hytrix command run async and returns the fallback error", t, func() {
		x := NewHystrixStringCommand("error", "fallbackError")

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
