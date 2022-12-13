// Copyright 2019 F5 Networks. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// various functions that take sequences of things and combine them into one thing
package fold

import (
	"sync"

	"github.com/weberr13/go-reducers/monoid"
)

const (
	DefaultConcurrentWorkers = 32
)

func drainToOne(in <-chan monoid.CommutativeMonoid, out chan<- monoid.CommutativeMonoid) {
	for a := range in {
		select {
		case b := <-in:
			out <- a.Two(b)
		default:
			out <- a
			return
		}
	}
}

func folder(c chan monoid.CommutativeMonoid, done <-chan struct{}, i monoid.Identity, wg *sync.WaitGroup) {
	myWg := &sync.WaitGroup{}
	loopback := make(chan monoid.CommutativeMonoid, 1024)
	myWg.Add(1)
	go func() {
		defer myWg.Done()
		for l := range loopback {
			c <- l
		}
	}()
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
	one := false
loop:
	for {
		select {
		case a := <-c:
			one = true
			select {
			case b := <-c:
				loopback <- a.Two(b)
			case <-done:
				loopback <- a
				break loop
			}
		case <-done:
			if !one {
				loopback <- i()
				return
			}
			break loop
		}
	}
	drainToOne(c, loopback)
	close(loopback)
	myWg.Wait()
	drainToOne(c, c)
}

// FoldSlice of commutnative monoids (default threadpool)
func FoldSlice(in []monoid.CommutativeMonoid, i monoid.Identity) monoid.CommutativeMonoid {
	return FoldSliceN(in, i, DefaultConcurrentWorkers)
}

// FoldSliceN wide of commutnative monoids
func FoldSliceN(in []monoid.CommutativeMonoid, i monoid.Identity, numWorkers int) monoid.CommutativeMonoid {
	return FoldSourceN(func(lazy chan<- monoid.CommutativeMonoid) {
		for i := range in {
			lazy <- in[i]
		}
	}, i, numWorkers)
}

// FoldChan elements till the channel is closed() (default threadpool)
func FoldChan(in chan monoid.CommutativeMonoid, i monoid.Identity) monoid.CommutativeMonoid {
	return FoldChanN(in, i, DefaultConcurrentWorkers)
}

// FoldChanN elements till the channel is closed() N wide
func FoldChanN(in chan monoid.CommutativeMonoid, i monoid.Identity, numWorkers int) monoid.CommutativeMonoid {
	return FoldSourceN(func(lazy chan<- monoid.CommutativeMonoid) {
		for c := range in {
			lazy <- c
		}
	}, i, numWorkers)
}

type SourceData func(chan<- monoid.CommutativeMonoid)

// FoldSource data from a function that feeds a channel and exists (default threadpool)
func FoldSource(f SourceData, i monoid.Identity) monoid.CommutativeMonoid {
	return FoldSourceN(f, i, DefaultConcurrentWorkers)
}

// FoldSourceN data from a funciton that feeds a channel and exits with given threadpool size
func FoldSourceN(f SourceData, i monoid.Identity, numWorkers int) monoid.CommutativeMonoid {
	if numWorkers < 1 {
		numWorkers = 1
	}
	if numWorkers > 512 {
		numWorkers = 512
	}
	lazy := make(chan monoid.CommutativeMonoid, 1024)
	done := make(chan struct{})
	wg := &sync.WaitGroup{}
	for j := 0; j < numWorkers; j++ {
		wg.Add(1)
		go folder(lazy, done, i, wg)
	}
	f(lazy)
	close(done)
	wg.Wait()
	drainToOne(lazy, lazy)
	return <-lazy
}
