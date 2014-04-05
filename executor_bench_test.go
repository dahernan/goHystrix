// Basic benchmark
// go test -test.run="^bench$" -bench=.
//
// BenchmarkRun		50000000	  67.1 ns/op
// BenchmarkExecute	  500000	7835.0 ns/op
//
package goHystrix

import (
	"testing"
	"time"
)

var (
	command = NewCommand(&SimpleCommand{})
)

type SimpleCommand struct {
}

func (rc *SimpleCommand) Name() string           { return "SimpleCommand" }
func (rc *SimpleCommand) Group() string          { return "testGroup" }
func (rc *SimpleCommand) Timeout() time.Duration { return 1 * time.Second }
func (rc *SimpleCommand) Run() (interface{}, error) {
	return "ok", nil
}

func benchmarkRunN(b *testing.B) {
	for n := 0; n < b.N; n++ {
		command.Run()
	}
}

func benchmarkExecuteN(b *testing.B) {
	for n := 0; n < b.N; n++ {
		command.Execute()
	}
}

func BenchmarkRun(b *testing.B)     { benchmarkRunN(b) }
func BenchmarkExecute(b *testing.B) { benchmarkExecuteN(b) }
