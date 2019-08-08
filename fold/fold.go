package fold

import (
	"sync"
	"github.com/weberr13/go-reducers/monoid"
)

const (
	DefaultThreads = 128
)

func folder(c chan monoid.CommunativeMonoid, done chan struct{}, wg *sync.WaitGroup) {
	var a, b monoid.CommunativeMonoid

	defer wg.Done()
loop:
	for a = range c {
		select {
		case b = <-c:
			c <- a.Two(b)
			a = nil
		case <- done:
			c <- a.One()
			break loop
		}
	}
	// drain
	for {
		select {
		case a = <-c: 
			select {
			case b = <- c:
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

func FoldSlice(in []monoid.CommunativeMonoid) monoid.CommunativeMonoid {
	return FoldSliceN(in, DefaultThreads)
}

func FoldSliceN(in []monoid.CommunativeMonoid, threads int) monoid.CommunativeMonoid {
	lazy := make(chan monoid.CommunativeMonoid, threads)
	done := make(chan struct{})
	wg := &sync.WaitGroup{}
	for j := 0 ; j < threads ; j++ {
		wg.Add(1)
		go folder(lazy, done, wg)
	}
	for i := 0 ; i < len(in) ; i++ {
		lazy <- in[i]
	}
	close(done)
	wg.Wait()
	return <- lazy
}