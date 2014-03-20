package goHystrix

import (
	"fmt"
	"time"
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
			metric := NewMetric("group1", "test")
			Convey("The metric holder should contain that metric", func() {
				value, ok := metrics.Get("group1", "test")
				So(value, ShouldEqual, metric)
				So(ok, ShouldBeTrue)

			})

		})
	})
}

func TestCountersSuccess(t *testing.T) {
	Convey("Metric stores the success counter", t, func() {
		Metrics()
		metric := NewMetric("group1", "test")

		metric.Success(1)
		metric.Success(2)

		c1 := metric.HealthCounts()

		s := c1.Success

		So(s, ShouldEqual, 2)
		So(c1.Failures, ShouldEqual, 0)
		So(c1.Timeouts, ShouldEqual, 0)
		So(c1.Fallback, ShouldEqual, 0)
		So(c1.FallbackErrors, ShouldEqual, 0)

		metric.Success(3)

		c2 := metric.HealthCounts()

		So(c1.Success, ShouldEqual, 2)

		So(c2.Success, ShouldEqual, 3)
		So(c2.Failures, ShouldEqual, 0)
		So(c2.Timeouts, ShouldEqual, 0)
		So(c2.Fallback, ShouldEqual, 0)
		So(c2.FallbackErrors, ShouldEqual, 0)

	})
}

func TestCounters(t *testing.T) {
	Convey("Metric stores the others counters", t, func() {
		Metrics()
		metric := NewMetric("group3", "test3")

		metric.Success(1)
		metric.Success(2)
		metric.Success(3)
		metric.Success(4)
		metric.Fail()
		metric.Fail()
		metric.Fail()
		metric.Fallback()
		metric.Fallback()
		metric.FallbackError()
		metric.FallbackError()
		metric.FallbackError()
		metric.Timeout()
		metric.Timeout()
		metric.Timeout()
		metric.Timeout()

		c1 := metric.HealthCounts()

		So(c1.Success, ShouldEqual, 4)
		So(c1.Failures, ShouldEqual, 7)
		So(c1.Timeouts, ShouldEqual, 4)
		So(c1.Fallback, ShouldEqual, 2)
		So(c1.FallbackErrors, ShouldEqual, 3)
		So(c1.Total, ShouldEqual, 11)
		So(c1.ErrorPercentage, ShouldEqual, 63.63636363636363)

		metric.Fail()
		metric.Success(5)
		metric.Fallback()
		metric.FallbackError()
		metric.Timeout()

		c2 := metric.HealthCounts()

		So(c2.Success, ShouldEqual, 5)
		So(c2.Failures, ShouldEqual, 9)
		So(c2.Timeouts, ShouldEqual, 5)
		So(c2.Fallback, ShouldEqual, 3)
		So(c2.FallbackErrors, ShouldEqual, 4)
		So(c2.Total, ShouldEqual, 14)
		So(c2.ErrorPercentage, ShouldEqual, 64.28571428571429)

	})
}

func TestRollingsCounters(t *testing.T) {
	Convey("Metric stores the counters in buckets for rolling the counters", t, func() {
		Metrics()
		metric := NewMetricWithParams("group2", "test", 4, 10)
		fmt.Println("== metric.Success(1)")
		metric.Success(1)
		metric.Success(1)
		metric.Success(1)
		metric.Fail()
		metric.Fallback()
		metric.FallbackError()
		metric.Timeout()
		time.Sleep(3 * time.Second)

		fmt.Println("== metric.Success(2)")
		metric.Success(2)
		metric.Success(2)
		time.Sleep(1 * time.Second)
		fmt.Println("== metric.Success(3)")
		metric.Fail()
		metric.Fail()
		metric.Fallback()
		metric.FallbackError()
		metric.Timeout()

		metric.Success(3)
		time.Sleep(1 * time.Second)
		fmt.Println("== metric.Success(4)")
		metric.Success(4)
		metric.Fail()
		metric.Fail()
		metric.Fail()
		metric.Fallback()
		metric.FallbackError()
		metric.Timeout()

		c1 := metric.HealthCounts()

		fmt.Println("== c1", c1)

		So(c1.Success, ShouldEqual, 4)
		So(c1.Failures, ShouldEqual, 7)
		So(c1.Timeouts, ShouldEqual, 2)
		So(c1.Fallback, ShouldEqual, 2)
		So(c1.FallbackErrors, ShouldEqual, 2)
		So(c1.Total, ShouldEqual, 11)
		So(c1.ErrorPercentage, ShouldEqual, 63.63636363636363)

	})
}

func TestMetricsKeepMessuaresSample(t *testing.T) {
	Convey("Keep the stats of the duration for the successful results", t, func() {
		Metrics()
		metric := NewMetricWithParams("group123", "test", 4, 20)
		metric.Success(5)
		metric.Success(1)
		metric.Success(9)
		metric.Success(2)
		metric.Success(5)
		metric.Success(8)

		So(metric.Stats().Max(), ShouldEqual, 9)
		So(metric.Stats().Min(), ShouldEqual, 1)
		// flacky because is calculanting async
		//So(metric.Stats().Mean(), ShouldEqual, 5)
		//So(metric.Stats().Count(), ShouldEqual, 6)
		//So(metric.Stats().Variance(), ShouldEqual, 8.333333333333334)

	})
}
