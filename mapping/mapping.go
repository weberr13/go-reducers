// Contains various functions that apply a function across sequences of things
package mapping

import (
	"sync"
)

// Copyable object must return a deep copy of itself with Copy
type Copyable interface {
	Copy() Copyable
}

// Sequencer a channel of things and close the channel once they are finished
type Sequencer func(chan<- Copyable)

// Transformer from a Copyable to a new thing
type Transformer func(Copyable) interface{}

// DefaultThreads for distributing tasks
const DefaultThreads = 128

// ForEach thing in input apply f and put it in out but unlike Map order is not preserved
// Closes out when complete
func ForEach(in Sequencer, out chan<- interface{}, f Transformer) {
	ForEachN(in, out, f, DefaultThreads)
}

// ForEachN thing in input apply f and put it in out but unlike Map order is not preserved
// runs with a pool of N instead of the default, Closes out when complete
func ForEachN(in Sequencer, out chan<- interface{}, f Transformer, threads int) {
	if threads < 1 {
		threads = 1
	}
	if threads > 1024 {
		threads = 1024
	}
	wg := &sync.WaitGroup{}
	c := make(chan Copyable, 1024)
	go in(c)
	for i := 0 ; i < threads ; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for d := range c {
				dc := d.Copy()
				out <- f(dc)
			}
		}()
	}
	wg.Wait()
	close(out)
}