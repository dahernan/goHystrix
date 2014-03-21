Implementation in Go (aka golang) of some of the stability patterns, like Circuit Breaker.

* Inspired by Nexflix Hystrix https://github.com/Netflix/Hystrix
* And the book Release it! http://www.amazon.co.uk/Release-Production-Ready-Software-Pragmatic-Programmers-ebook/dp/B00A32NXZO/


How to use
----------

### You need to implement goHystrix.Interface

```go
type Interface interface {
	Run() (interface{}, error)
	Fallback() (interface{}, error)
	Timeout() time.Duration
	Name() string
	Group() string
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

func (c *MyStringCommand) Name() string {
	return "stringMessage"
}

func (c *MyStringCommand) Group() string {
	return "stringGroup"
}

func (c *MyStringCommand) Timeout() time.Duration {
	return 3 * time.Millisecond
}

func (c *MyStringCommand) Run() (interface{}, error) {
	return c.message, nil
}

func (c *MyStringCommand) Fallback() (interface{}, error) {
	return "FALLBACK", nil
}

func TestString(t *testing.T) {

	// Sync execution
	command := goHystrix.NewCommand(&MyStringCommand{"helloooooooo"})
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
errorPercetageThreshold - 50.0 - If number_of_errors / total_calls * 100 > 50.0 the circuit will be open
minimumNumberOfRequest - if total_calls < 20 the circuit will be close
numberOfSecondsToStore - 20 seconds (for health counts you only evaluate the last 20 seconds of calls)
numberOfSamplesToStore - 50 values (you store the duration of the 50 successful calls using reservoir sampling)
```

### You can customize the default values when you create the command
```go

// errorPercetageThreshold - 60.0
// minimumNumberOfRequest - 3
// numberOfSecondsToStore - 5
// numberOfSamplesToStore - 10
NewCommandWithParams(&AnotherCommand{}, 60.0, 3, 5, 10)
```

### Exposes all circuits information by http in "http://<host>/debug/circuits" in JSON format
```go
import	_ "github.com/dahernan/goHystrix/httpexp"
```



