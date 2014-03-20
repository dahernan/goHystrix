package goHystrix

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCircuitsHolder(t *testing.T) {
	Convey("Circuits Holder simple Get", t, func() {
		circuits := NewCircuitsHolder()
		Convey("When Get is called in an empty map", func() {
			value, ok := circuits.Get("testGroup", "testKey")
			Convey("The result should be nil", func() {
				So(value, ShouldBeNil)
			})
			Convey("The second returned value should be false", func() {
				So(ok, ShouldBeFalse)
			})

		})
	})

	Convey("Set and Get logic", t, func() {
		circuits := NewCircuitsHolder()
		Convey("When the circuits is filled with multiple values", func() {

			m1 := &CircuitBreaker{name: "1"}
			m2 := &CircuitBreaker{name: "2"}
			m3 := &CircuitBreaker{name: "3"}
			m4 := &CircuitBreaker{name: "4"}

			circuits.Set("testGroup1", "testKey1", m1)
			circuits.Set("testGroup1", "testKey2", m2)

			circuits.Set("testGroup2", "testKey1", m3)
			circuits.Set("testGroup2", "testKey2", m4)

			Convey("Get return that values back", func() {
				value, ok := circuits.Get("testGroup1", "testKey1")
				So(value, ShouldEqual, m1)
				So(ok, ShouldBeTrue)

				value, ok = circuits.Get("testGroup1", "testKey2")
				So(value, ShouldEqual, m2)
				So(ok, ShouldBeTrue)

				value, ok = circuits.Get("testGroup2", "testKey1")
				So(value, ShouldEqual, m3)
				So(ok, ShouldBeTrue)

				value, ok = circuits.Get("testGroup2", "testKey2")
				So(value, ShouldEqual, m4)
				So(ok, ShouldBeTrue)

			})

		})
	})

}

func TestNewCircuit(t *testing.T) {
	Convey("New Circuit register itself into the Circuit Holder", t, func() {
		NewCircuitsHolder()

		m1 := NewCircuitNoParams("testGroup1", "testKey1")
		m2 := NewCircuitNoParams("testGroup1", "testKey2")
		m3 := NewCircuitNoParams("testGroup2", "testKey1")
		m4 := NewCircuitNoParams("testGroup2", "testKey2")

		fmt.Println(Circuits())
		fmt.Println(Circuits().ToJSON())

		value, ok := Circuits().Get("testGroup1", "testKey1")
		So(value, ShouldEqual, m1)
		So(ok, ShouldBeTrue)

		value, ok = Circuits().Get("testGroup1", "testKey2")
		So(value, ShouldEqual, m2)
		So(ok, ShouldBeTrue)

		value, ok = Circuits().Get("testGroup2", "testKey1")
		So(value, ShouldEqual, m3)
		So(ok, ShouldBeTrue)

		value, ok = Circuits().Get("testGroup2", "testKey2")
		So(value, ShouldEqual, m4)
		So(ok, ShouldBeTrue)

		fmt.Println("########################")
		fmt.Println(Circuits().ToJSON())

	})

}
