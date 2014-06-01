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
	command := NewCommand("stringMessage", "stringGroup", &MyStringCommand{"helloooooooo"})
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
	fmt.Println("Mean: ", command.Metric().Stats().Mean())

	fmt.Println("All the circuits in JSON format: ", Circuits().ToJSON())

}

func TestStringWithOptions(t *testing.T) {

	// Sync execution
	command := NewCommandWithOptions("stringMessage", "stringGroup", &MyStringCommand{"helloooooooo"}, CommandOptions{
		ErrorsThreshold:        60.0,
		MinimumNumberOfRequest: 3,
		NumberOfSecondsToStore: 5,
		NumberOfSamplesToStore: 10,
	})
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
	fmt.Println("Mean: ", command.Metric().Stats().Mean())

	fmt.Println("All the circuits in JSON format: ", Circuits().ToJSON())

}
