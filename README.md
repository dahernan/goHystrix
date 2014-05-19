Implementation in Go (aka golang) of some of the stability patterns, like Circuit Breaker.

* Inspired by Nexflix Hystrix https://github.com/Netflix/Hystrix
* And the book Release it! http://www.amazon.co.uk/Release-Production-Ready-Software-Pragmatic-Programmers-ebook/dp/B00A32NXZO/


How to use
----------

### You need to implement goHystrix.Interface

```go
type Interface interface {
	Run() (interface{}, error)
	Timeout() time.Duration
	Name() string
	Group() string
}
```

There is also a `goHystrix.FallbackInterface`, if you need a fallback function:

```go
type FallbackInterface interface {
	Interface
	Fallback() (interface{}, error)
}
```

### Basic command with a String
```go
import (
	"fmt"
	"github.com/dahernan/goHystrix"
	"testing"
	"time"
)

type MyStringCommand struct {
	message string
}

// name of your command 
func (c *MyStringCommand) Name() string {
	return "stringMessageCommand"
}

// group of the command, it's good to keep related commands together 
func (c *MyStringCommand) Group() string {
	return "stringGroup"
}

// timeout for the command Run
func (c *MyStringCommand) Timeout() time.Duration {
	return 3 * time.Millisecond
}

// This is the normal method to execute (circuit is close) 
func (c *MyStringCommand) Run() (interface{}, error) {
	return c.message, nil
}

// This is the method to execute in case the circuit is open
func (c *MyStringCommand) Fallback() (interface{}, error) {
	return "FALLBACK", nil
}

func TestString(t *testing.T) {
	// creates a new command
	command := goHystrix.NewCommand(&MyStringCommand{"helloooooooo"})
	
	// Sync execution
	value, err := command.Execute()

	fmt.Println("Sync call ---- ")
	fmt.Println("Value: ", value)
	fmt.Println("Error: ", err)

	// Async execution

	valueChan, errorChan := command.Queue()

	fmt.Println("Async call ---- ")
	select {
	case value = <-valueChan:
		fmt.Println("Value: ", value)
	case err = <-errorChan:
		fmt.Println("Error: ", err)
	}

	fmt.Println("Succesfull Calls ", command.HealthCounts().Success)
	fmt.Println("Mean: ", command.Stats().Mean())

	fmt.Println("All the circuits in JSON format: ", goHystrix.Circuits().ToJSON())


}

```

### Default circuit values when you create a command
```
errorPercetageThreshold - 50.0 - If (number_of_errors / total_calls * 100) > 50.0 the circuit will be open
minimumNumberOfRequest - if total_calls < 20 the circuit will be close
numberOfSecondsToStore - 20 seconds (for health counts you only evaluate the last 20 seconds of calls)
numberOfSamplesToStore - 50 values (you store the duration of 50 successful calls using reservoir sampling)
```

### You can customize the default values when you create the command
```go

// ErrorPercetageThreshold - 60.0
// MinimumNumberOfRequest - 3
// NumberOfSecondsToStore - 5
// NumberOfSamplesToStore - 10
goHystrix.NewCommandWithOptions(&MyStringCommand{"helloooooooo"}, goHystrix.CommandOptions{
		ErrorsThreshold:        60.0,
		MinimumNumberOfRequest: 3,
		NumberOfSecondsToStore: 5,
		NumberOfSamplesToStore: 10,
	})

```

### Exposes all circuits information by http in JSON format
```go
import	_ "github.com/dahernan/goHystrix/httpexp"
```
GET - http://host/debug/circuits  


### Exposes the metrics using statds

```go
// host and prefix for statds server, and send the gauges of the state of the circuits every 3 Seconds
goHystrix.UseStatsd("0.0.0.0:8125", "myprefix", 3*time.Second)
```




