Implementation in Go (aka golang) of some of the stability patterns, like Circuit Breaker.

* Inspired by Nexflix Hystrix https://github.com/Netflix/Hystrix
* And the book Release it! http://www.amazon.co.uk/Release-Production-Ready-Software-Pragmatic-Programmers-ebook/dp/B00A32NXZO/

You can also check [breaker](https://github.com/dahernan/breaker) my other implementaion of the Circuit Breaker, simpler and with not dependencies: https://github.com/dahernan/breaker


How to use
----------

### You need to implement goHystrix.Interface

```go
type Interface interface {
	Run() (interface{}, error)
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
	command := goHystrix.NewCommand("commandName", "commandGroup", &MyStringCommand{"helloooooooo"})
	
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
ErrorPercetageThreshold - 50.0 - If (number_of_errors / total_calls * 100) > 50.0 the circuit will be open
MinimumNumberOfRequest - if total_calls < 20 the circuit will be close
NumberOfSecondsToStore - 20 seconds (for health counts you only evaluate the last 20 seconds of calls)
NumberOfSamplesToStore - 50 values (you store the duration of 50 successful calls using reservoir sampling)
Timeout - 2 * time.Seconds
```

### You can customize the default values when you create the command
```go

// ErrorPercetageThreshold - 60.0
// MinimumNumberOfRequest - 3
// NumberOfSecondsToStore - 5
// NumberOfSamplesToStore - 10
// Timeout - 10 * time.Second
goHystrix.NewCommandWithOptions("commandName", "commandGroup", &MyStringCommand{"helloooooooo"}, goHystrix.CommandOptions{
		ErrorsThreshold:        60.0,
		MinimumNumberOfRequest: 3,
		NumberOfSecondsToStore: 5,
		NumberOfSamplesToStore: 10,
		Timeout:                10 * time.Second,
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




