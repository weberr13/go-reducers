package fold

import (
	"sync"

	"github.com/weberr13/go-reducers/monoid"
)

const (
	DefaultThreads = 128
)

func folder(c chan monoid.CommutativeMonoid, done chan struct{}, wg *sync.WaitGroup) {
	var a, b monoid.CommutativeMonoid

	defer wg.Done()
loop:
	for a = range c {
		select {
		case b = <-c:
			c <- a.Two(b)
		case <-done:
			c <- a.One()
			break loop
		}
	}
	// drain
	for {
		select {
		case a = <-c:
			select {
			case b = <-c:
				c <- a.Two(b)
			default:
				c <- a.One()
				return
			}
		default:
			return
		}
	}
}

func FoldSlice(in []monoid.CommutativeMonoid) monoid.CommutativeMonoid {
	return FoldSliceN(in, DefaultThreads)
}

func FoldSliceN(in []monoid.CommutativeMonoid, threads int) monoid.CommutativeMonoid {
	lazy := make(chan monoid.CommutativeMonoid, threads)
	done := make(chan struct{})
	wg := &sync.WaitGroup{}
	for j := 0; j < threads; j++ {
		wg.Add(1)
		go folder(lazy, done, wg)
	}
	for i := 0; i < len(in); i++ {
		lazy <- in[i]
	}
	close(done)
	wg.Wait()
	// The following feels hacky but there is a race where multiple folders could each get 1 element and all
	// believe that they are the end of the lazy seq.  In that case we need a last merge.
	var c monoid.CommutativeMonoid
	for {
		select {
		case cp := <-lazy:
			if c == nil {
				c = cp
			} else {
				c = cp.Two(c)
			}
		default:
			return c
		}
	}
}
