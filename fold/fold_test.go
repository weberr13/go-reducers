// Copyright 2019 F5 Networks. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fold

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/weberr13/go-reducers/mapping"
	"github.com/weberr13/go-reducers/monoid"
)

type StringSet struct {
	mine map[string]struct{}
}

func IdenityStringSet() monoid.CommutativeMonoid {
	return &StringSet{
		mine: make(map[string]struct{}),
	}
}

func (f StringSet) Two(b monoid.CommutativeMonoid) monoid.CommutativeMonoid {
	realB, ok := b.(*StringSet)
	if !ok {
		panic("Runtime error!  Dynamic cast failure for monoid.CommutativeMonoid -> StringSet")
	}
	n := make(map[string]struct{})
	for k := range f.mine {
		n[k] = struct{}{}
	}
	for k := range realB.mine {
		n[k] = struct{}{}
	}
	return &StringSet{
		mine: n,
	}
}

type CopyableString string

func (s CopyableString) Copy() mapping.Copyable {
	c := s
	return c
}

type WordCount struct {
	mine map[CopyableString]int
}

func IdentityWordCount() monoid.CommutativeMonoid {
	return &WordCount{
		mine: make(map[CopyableString]int),
	}
}

func NewWordCount(c mapping.Copyable) interface{} {
	w := &WordCount{
		mine: make(map[CopyableString]int),
	}
	cs := c.(CopyableString)
	s := string(cs)
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	s = reg.ReplaceAllString(s, " ")
	s = strings.ToLower(s)
	sn := strings.Fields(s)

	for _, v := range sn {
		vc := CopyableString(v)
		if _, ok := w.mine[vc]; ok {
			w.mine[vc]++
		} else {
			w.mine[vc] = 1
		}
	}
	return w
}

func (f WordCount) Two(b monoid.CommutativeMonoid) monoid.CommutativeMonoid {
	realB, ok := b.(*WordCount)
	if !ok {
		panic("Runtime error!  Dynamic cast failure for monoid.CommutativeMonoid -> WordCount")
	}
	n := make(map[CopyableString]int)
	for k, v := range f.mine {
		n[k] = v
	}
	for k, v := range realB.mine {
		if _, ok := n[k]; ok {
			n[k] = n[k] + v
		} else {
			n[k] = v
		}
	}
	return &WordCount{
		mine: n,
	}
}

func WarAndPeaceCount(b *testing.B, forDepth int, foldDepth int, readlines int, linebundle int) {
	filename := filepath.Join("testdata", "WarAndPeace_LeoTolstoy.txt.gz")
	file, err := os.Open(filename)
	if err != nil {
		b.Fatal(err)
	}
	gz, err := gzip.NewReader(file)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	defer gz.Close()
	count := 0
	scanner := bufio.NewScanner(gz)
	input := []string{}
	linecount := 1
loop:
	for scanner.Scan() {
		if count > readlines {
			break loop
		}
		if linecount < linebundle && len(input) > 0 {
			input[len(input)-1] += " " + scanner.Text()
			linecount++
		} else {
			input = append(input, scanner.Text())
			linecount = 1
		}
		count++
	}
	expectedFile := filepath.Join("testdata", fmt.Sprintf("WarAndPeaceResults.%d.json", readlines))
	refdata, _ := ioutil.ReadFile(expectedFile)
	refmap := make(map[string]interface{})
	err = json.Unmarshal(refdata, &refmap)
	if err != nil {
		refmap = nil
	} else {
		refmap = nil
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		creator := func(c chan<- monoid.CommutativeMonoid) {
			wg := &sync.WaitGroup{}
			out := make(chan interface{}, 1024)
			wg.Add(1)
			go func() {
				defer wg.Done()
				mapping.ForEachN(
					func(c chan<- mapping.Copyable) {
						for _, line := range input {
							c <- CopyableString(line)
						}
						close(c)
					},
					out,
					NewWordCount,
					forDepth,
				)
			}()
			for raw := range out {
				s, ok := raw.(monoid.CommutativeMonoid)
				if !ok {
					b.Fatal("could not cast result of ForEach")
				}
				c <- s
			}
			wg.Wait()
		}
		c := FoldSourceN(creator, IdentityWordCount, foldDepth)
		cr, ok := c.(*WordCount)
		if !ok {
			b.FailNow()
		}
		if refmap != nil {
			if !reflect.DeepEqual(refmap, cr.mine) {
				b.Fatalf("error in results for %d %d %d", forDepth, foldDepth, readlines)
			}
		} else {
			data, err := json.Marshal(cr.mine)
			if err != nil {
				b.Fatal(err)
			}
			err = ioutil.WriteFile(expectedFile, data, 0644)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkWordCount(b *testing.B) {
	fors := []int{1, 16, 24}
	folds := []int{1, 2, 4, 6, 8}
	bundleSize := []int{1, 10, 100}
	for _, bs := range bundleSize {
		for _, fl := range folds {
			for _, fo := range fors {
				b.Run(fmt.Sprintf("Fold %d deep Map %d deep with %d lines per Map", fl, fo, bs), func(b *testing.B) {
					WarAndPeaceCount(b, fo, fl, 10000, bs)
				})
			}
		}
	}

}

func TestWordCount(t *testing.T) {
	Convey("short word count example", t, func() {
		testData := []CopyableString{
			"The The is a great band, yes is also a great one.",
			"the person who wrote the other line has terrible taste",
		}

		creator := func(c chan<- monoid.CommutativeMonoid) {
			out := make(chan interface{}, 100)
			mapping.ForEach(
				func(c chan<- mapping.Copyable) {
					for _, v := range testData {
						c <- v
					}
					close(c)
				},
				out,
				NewWordCount,
			)
			for raw := range out {
				s, ok := raw.(monoid.CommutativeMonoid)
				if !ok {
					panic("could not cast result of ForEach")
				}
				c <- s
			}
		}
		c := FoldSource(creator, IdentityWordCount)
		cr, ok := c.(*WordCount)
		So(ok, ShouldBeTrue)
		So(cr.mine, ShouldResemble, map[CopyableString]int{
			"a":        2,
			"also":     1,
			"band":     1,
			"great":    2,
			"has":      1,
			"is":       2,
			"line":     1,
			"one":      1,
			"other":    1,
			"person":   1,
			"taste":    1,
			"terrible": 1,
			"the":      4,
			"who":      1,
			"wrote":    1,
			"yes":      1,
		})
	})
}

func TestExampleFoldSlice(t *testing.T) {
	Convey("simple set union fold", t, func() {
		s := []monoid.CommutativeMonoid{
			&StringSet{map[string]struct{}{"foo": struct{}{}}},
			&StringSet{map[string]struct{}{"bar": struct{}{}}},
			&StringSet{map[string]struct{}{"baz": struct{}{}}},
			&StringSet{map[string]struct{}{"foo": struct{}{}}},
		}
		c := FoldSlice(s, IdenityStringSet)
		cr, ok := c.(*StringSet)
		So(ok, ShouldBeTrue)
		So(cr, ShouldResemble, &StringSet{map[string]struct{}{
			"foo": struct{}{},
			"bar": struct{}{},
			"baz": struct{}{},
		}})
	})
	Convey("empty set", t, func() {
		s := []monoid.CommutativeMonoid{}
		c := FoldSlice(s, IdenityStringSet)
		cr, ok := c.(*StringSet)
		So(ok, ShouldBeTrue)
		So(cr, ShouldResemble, &StringSet{map[string]struct{}{}})
	})

}

func TestExampleFoldChan(t *testing.T) {
	Convey("simple set union fold", t, func() {
		s := make(chan monoid.CommutativeMonoid, 10)
		go func() {
			s <- &StringSet{map[string]struct{}{"foo": struct{}{}}}
			s <- &StringSet{map[string]struct{}{"bar": struct{}{}}}
			s <- &StringSet{map[string]struct{}{"baz": struct{}{}}}
			s <- &StringSet{map[string]struct{}{"foo": struct{}{}}}
			close(s)
		}()

		c := FoldChan(s, IdenityStringSet)
		cr, ok := c.(*StringSet)
		So(ok, ShouldBeTrue)
		So(cr, ShouldResemble, &StringSet{map[string]struct{}{
			"foo": struct{}{},
			"bar": struct{}{},
			"baz": struct{}{},
		}})
	})
	Convey("simple set union fold, lower boundary check", t, func() {
		s := make(chan monoid.CommutativeMonoid, 10)
		go func() {
			s <- &StringSet{map[string]struct{}{"foo": struct{}{}}}
			s <- &StringSet{map[string]struct{}{"bar": struct{}{}}}
			s <- &StringSet{map[string]struct{}{"baz": struct{}{}}}
			s <- &StringSet{map[string]struct{}{"foo": struct{}{}}}
			close(s)
		}()

		c := FoldChanN(s, IdenityStringSet, -1)
		cr, ok := c.(*StringSet)
		So(ok, ShouldBeTrue)
		So(cr, ShouldResemble, &StringSet{map[string]struct{}{
			"foo": struct{}{},
			"bar": struct{}{},
			"baz": struct{}{},
		}})
	})
	Convey("simple set union fold, upper boundary check", t, func() {
		s := make(chan monoid.CommutativeMonoid, 10)
		go func() {
			s <- &StringSet{map[string]struct{}{"foo": struct{}{}}}
			s <- &StringSet{map[string]struct{}{"bar": struct{}{}}}
			s <- &StringSet{map[string]struct{}{"baz": struct{}{}}}
			s <- &StringSet{map[string]struct{}{"foo": struct{}{}}}
			close(s)
		}()

		c := FoldChanN(s, IdenityStringSet, 513)
		cr, ok := c.(*StringSet)
		So(ok, ShouldBeTrue)
		So(cr, ShouldResemble, &StringSet{map[string]struct{}{
			"foo": struct{}{},
			"bar": struct{}{},
			"baz": struct{}{},
		}})
	})
}
