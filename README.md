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