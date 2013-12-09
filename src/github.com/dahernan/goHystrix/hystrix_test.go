package goHystrix

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

type HystrixStringCommand struct {
	state string
	*HystrixExecutor
}

func NewHystrixStringCommand(state string) *HystrixStringCommand {
	command := &HystrixStringCommand{}
	executor := NewHystrixExecutor(command)
	command.state = state
	command.HystrixExecutor = executor
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

func TestRunsOk(t *testing.T) {

	Convey("Hytrix command runs properly", t, func() {
		x := NewHystrixStringCommand("ok")

		Convey("When Run is executed", func() {

			result, err := x.Execute()

			Convey("The result should be the string value", func() {
				So(result, ShouldEqual, "hello hystrix world")
			})

			Convey("There is no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Hytrix command returns an error", t, func() {
		x := NewHystrixStringCommand("error")

		Convey("When Run is executed", func() {
			result, err := x.Execute()

			Convey("The result should be Nil", func() {
				So(result, ShouldBeNil)
			})

			Convey("There is an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Hytrix command returns error due to timeout", t, func() {
		x := NewHystrixStringCommand("timeout")

		Convey("When Run is executed", func() {
			result, err := x.Execute()

			Convey("The result should be Nil", func() {
				So(result, ShouldBeNil)
			})

			Convey("There is a timeout error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "ERROR: Timeout!!")
			})
		})
	})

	Convey("Hytrix command run async and returns ok", t, func() {
		x := NewHystrixStringCommand("ok")

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

			Convey("The result should be the string value", func() {
				So(result, ShouldEqual, "hello hystrix world")
			})
			Convey("There is no error", func() {
				So(err, ShouldBeNil)
			})

		})
	})

	Convey("Hytrix command run async and returns error", t, func() {
		x := NewHystrixStringCommand("error")

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

			Convey("The result should be the string value", func() {
				So(result, ShouldBeNil)
			})
			Convey("There is an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Hytrix command run async and returns a timeout error", t, func() {
		x := NewHystrixStringCommand("timeout")

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

			Convey("The result should be the string value", func() {
				So(result, ShouldBeNil)
			})
			Convey("There is an timeout error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "ERROR: Timeout!!")
			})
		})
	})

}
