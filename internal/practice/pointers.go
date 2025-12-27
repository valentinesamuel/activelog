package practice

import "fmt"

type Counter struct {
	count int
}

func (c Counter) IncrementValue() {
	c.count++
	fmt.Printf("Inside IncrementValue: %d\n", c.count)
}

func (c *Counter) IncrementPointer() {
	c.count++
	fmt.Printf("Inside IncrementPointer: %d\n", c.count)
}

func DemoPointers() {
	counter1 := Counter{count: 0}
	counter1.IncrementValue()
	fmt.Printf("After value increment %d\n", counter1.count)

	counter2 := Counter{count: 0}
	counter2.IncrementPointer()
	fmt.Printf("After pointer incerment: %d\n", counter2.count)
}
