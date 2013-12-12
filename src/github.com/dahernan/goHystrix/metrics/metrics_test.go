package metrics

import (
	//"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	//"time"
)

func TestMetricsHolder(t *testing.T) {

	Convey("Metrics Holder simple Get", t, func() {
		metrics := NewMetricsHolder()
		Convey("When Get is called in an empty map", func() {
			value, ok := metrics.Get("testGroup", "testKey")
			Convey("The result should be nil", func() {
				So(value, ShouldBeNil)
			})
			Convey("The second returned value should be false", func() {
				So(ok, ShouldBeFalse)
			})

		})
	})

	Convey("Set and Get logic", t, func() {
		metrics := NewMetricsHolder()
		Convey("When the metrics is filled with multiple values", func() {

			m1 := &Metric{name: "1"}
			m2 := &Metric{name: "2"}
			m3 := &Metric{name: "3"}
			m4 := &Metric{name: "4"}

			metrics.Set("testGroup1", "testKey1", m1)
			metrics.Set("testGroup1", "testKey2", m2)

			metrics.Set("testGroup2", "testKey1", m3)
			metrics.Set("testGroup2", "testKey2", m4)

			Convey("Get return that values back", func() {
				value, ok := metrics.Get("testGroup1", "testKey1")
				So(value, ShouldEqual, m1)
				So(ok, ShouldBeTrue)

				value, ok = metrics.Get("testGroup1", "testKey2")
				So(value, ShouldEqual, m2)
				So(ok, ShouldBeTrue)

				value, ok = metrics.Get("testGroup2", "testKey1")
				So(value, ShouldEqual, m3)
				So(ok, ShouldBeTrue)

				value, ok = metrics.Get("testGroup2", "testKey2")
				So(value, ShouldEqual, m4)
				So(ok, ShouldBeTrue)

			})

		})
	})

}

func TestMetric(t *testing.T) {
	Convey("New Metrics are register in the metrics holder", t, func() {
		metrics := Metrics()
		Convey("When new metric is created", func() {
			metric := NewMetric("test", "group1")
			Convey("The metric holder should contain that metric", func() {
				value, ok := metrics.Get("group1", "test")
				So(value, ShouldEqual, metric)
				So(ok, ShouldBeTrue)

			})

		})
	})
}
