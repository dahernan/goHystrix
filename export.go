package goHystrix

import (
	"fmt"
	"github.com/dahernan/goHystrix/statsd"
	"time"
)

type StatsdExport struct {
	statsdClient *statsd.StatsdClient
	prefix       string
}

type NilExport struct {
}

func UseStatsd(host string, port int, prefix string) {
	statsdClient := statsd.New(host, port)
	export := NewStatsdExport(prefix, statsdClient)
	SetExporter(export)
}

func NewNilExport() MetricExport { return NilExport{} }

func (NilExport) Success(group string, name string, duration time.Duration) {}
func (NilExport) Fail(group string, name string)                            {}
func (NilExport) Fallback(group string, name string)                        {}
func (NilExport) FallbackError(group string, name string)                   {}
func (NilExport) Timeout(group string, name string)                         {}
func (NilExport) Panic(group string, name string)                           {}

func NewStatsdExport(prefix string, statsdClient *statsd.StatsdClient) MetricExport {
	return StatsdExport{statsdClient, prefix}
}

func (s StatsdExport) Success(group string, name string, duration time.Duration) {
	go func() {
		s.statsdClient.Increment(fmt.Sprintf("%s.%s.%s.success", s.prefix, group, name))
		ms := int64(duration / time.Millisecond)
		s.statsdClient.Timing(fmt.Sprintf("%s.%s.%s.duration", s.prefix, group, name), ms)
	}()
}

func (StatsdExport) Fail(group string, name string)          {}
func (StatsdExport) Fallback(group string, name string)      {}
func (StatsdExport) FallbackError(group string, name string) {}
func (StatsdExport) Timeout(group string, name string)       {}
func (StatsdExport) Panic(group string, name string)         {}
