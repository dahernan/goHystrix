package goHystrix

//import "github.com/dahernan/goHystrix"
import (
	"fmt"
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
	command := NewCommand(&MyStringCommand{"helloooooooo"})
	value, err := command.Execute()

	fmt.Println("Value: ", value)
	fmt.Println("Error: ", err)

	// Async execution
	valueChan, errorChan := command.Queue()

	select {
	case value = <-valueChan:
		fmt.Println("Value: ", value)
	case err = <-errorChan:
		fmt.Println("Error: ", err)
	}

	fmt.Println("Succesfull Calls ", command.HealthCounts().Success)

}
