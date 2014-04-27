package goHystrix

import (
	"fmt"
	"github.com/dahernan/goHystrix/statsd"
	"log"
	"time"
)

var (
	metricsExporter MetricExport
)

func init() {
	metricsExporter = NewNilExport()
}

func Exporter() MetricExport {
	return metricsExporter
}

func SetExporter(export MetricExport) {
	metricsExporter = export
}

type MetricExport interface {
	Success(group string, name string, duration time.Duration)
	Fail(group string, name string)
	Fallback(group string, name string)
	FallbackError(group string, name string)
	Timeout(group string, name string)
	Panic(group string, name string)
	State(circuits *CircuitHolder)
}

type StatsdExport struct {
	statsdClient statsd.Statter
	prefix       string
}

type NilExport struct {
}

func UseStatsd(address string, prefix string, dur time.Duration) {
	statsdClient, err := statsd.Dial("udp", address)
	if err != nil {
		log.Println("Error setting Statds for publishing the metrics: ", err)
		log.Println("Using NilExport for publishing the metrics")
		SetExporter(NilExport{})
		return
	}
	export := NewStatsdExport(statsdClient, prefix)
	SetExporter(export)

	statsdExport := export.(StatsdExport)
	// poll the state of the circuits
	go statsdExport.run(Circuits(), dur)
}

func NewNilExport() MetricExport { return NilExport{} }

func (NilExport) Success(group string, name string, duration time.Duration) {}
func (NilExport) Fail(group string, name string)                            {}
func (NilExport) Fallback(group string, name string)                        {}
func (NilExport) FallbackError(group string, name string)                   {}
func (NilExport) Timeout(group string, name string)                         {}
func (NilExport) Panic(group string, name string)                           {}
func (NilExport) State(circuits *CircuitHolder)                             {}

func NewStatsdExport(statsdClient statsd.Statter, prefix string) MetricExport {
	return StatsdExport{statsdClient, prefix}
}

func (s StatsdExport) Success(group string, name string, duration time.Duration) {
	go func() {
		s.statsdClient.Counter(1.0, fmt.Sprintf("%s.%s.%s.success", s.prefix, group, name), 1)
		//ms := int64(duration / time.Millisecond)
		s.statsdClient.Timing(1.0, fmt.Sprintf("%s.%s.%s.duration", s.prefix, group, name), duration)
	}()
}

func (s StatsdExport) Fail(group string, name string) {
	go func() {
		s.statsdClient.Counter(1.0, fmt.Sprintf("%s.%s.%s.fail", s.prefix, group, name), 1)
	}()
}
func (s StatsdExport) Fallback(group string, name string) {
	go func() {
		s.statsdClient.Counter(1.0, fmt.Sprintf("%s.%s.%s.fallback", s.prefix, group, name), 1)
	}()
}
func (s StatsdExport) FallbackError(group string, name string) {
	go func() {
		s.statsdClient.Counter(1.0, fmt.Sprintf("%s.%s.%s.fallbackError", s.prefix, group, name), 1)
	}()
}
func (s StatsdExport) Timeout(group string, name string) {
	go func() {
		s.statsdClient.Counter(1.0, fmt.Sprintf("%s.%s.%s.timeout", s.prefix, group, name), 1)
	}()
}
func (s StatsdExport) Panic(group string, name string) {
	go func() {
		s.statsdClient.Counter(1.0, fmt.Sprintf("%s.%s.%s.panic", s.prefix, group, name), 1)
	}()
}

func (s StatsdExport) State(holder *CircuitHolder) {
	// TODO: have a save way to iterate over the circuits without
	// knowing how is implemented
	holder.mutex.RLock()
	defer holder.mutex.RUnlock()
	for group, names := range holder.circuits {
		for name, circuit := range names {
			var state string
			open, _ := circuit.IsOpen()
			state = "0"
			if open {
				state = "1"
			}
			s.statsdClient.Gauge(1.0, fmt.Sprintf("%s.%s.%s.open", s.prefix, group, name), state)
		}
	}
}

func (s StatsdExport) run(holder *CircuitHolder, dur time.Duration) {
	for {
		select {
		case <-time.After(dur):
			s.State(holder)
		}
	}
}
