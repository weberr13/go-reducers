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

	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
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

// FoldSlice of commutnative monoids (default threadpool)
func FoldSlice(in []monoid.CommutativeMonoid) monoid.CommutativeMonoid {
	return FoldSliceN(in, DefaultThreads)
}

// FoldSliceN wide of commutnative monoids 
func FoldSliceN(in []monoid.CommutativeMonoid, threads int) monoid.CommutativeMonoid {
	return FoldSourceN(func(lazy chan monoid.CommutativeMonoid) {
		for i := range in {
			lazy <- in[i]
		}
	}, threads)
}

// FoldChan elements till the channel is closed() (default threadpool)
func FoldChan(in chan monoid.CommutativeMonoid) monoid.CommutativeMonoid {
	return FoldChanN(in, DefaultThreads)
}

// FoldChanN elements till the channel is closed() N wide
func FoldChanN(in chan monoid.CommutativeMonoid, threads int) monoid.CommutativeMonoid {
	return FoldSourceN(func(lazy chan monoid.CommutativeMonoid) {
		for c := range in {
			lazy <- c
		}
	}, threads)
}

type SourceData func(chan monoid.CommutativeMonoid)

// FoldSource data from a function that feeds a channel and exists (default threadpool)
func FoldSource(f SourceData) monoid.CommutativeMonoid {
	return FoldSourceN(f, DefaultThreads)
}

// FoldSourceN data from a funciton that feeds a channel and exits with given threadpool size
func FoldSourceN(f SourceData, threads int) monoid.CommutativeMonoid {
	lazy := make(chan monoid.CommutativeMonoid, threads)
	done := make(chan struct{})
	wg := &sync.WaitGroup{}
	for j := 0; j < threads; j++ {
		wg.Add(1)
		go folder(lazy, done, wg)
	}
	f(lazy)
	close(done)
	wg.Wait()
	folder(lazy, done, nil)
	return <- lazy
}