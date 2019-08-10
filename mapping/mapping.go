package mapping

import (
	"sync"
)

// Copyable object must return a deep copy of itself with Copy
type Copyable interface {
	Copy() Copyable
}

// Sequencers close the channel once they are finished
type Sequencer func(chan Copyable)

type Transformer func(Copyable) interface{}

func ForEach(in Sequencer, out chan interface{}, f Transformer) {
	wg := &sync.WaitGroup{}
	c := make(chan Copyable, 100)
	go in(c)
	for d := range c {
		wg.Add(1)
		dc := d.Copy()
		go func() {
			defer wg.Done()
			out <- f(dc)
		}()
	}
	wg.Wait()
	close(out)
}