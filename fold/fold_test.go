package fold

import (
	"strings"
	"testing"
	"regexp"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/weberr13/go-reducers/monoid"
	"github.com/weberr13/go-reducers/mapping"
)

type StringSet struct {
	mine map[string]struct{}
}

func IdenityStringSet() monoid.CommutativeMonoid {
	return &StringSet{
		mine: make(map[string]struct{}),
	}
}

func (f StringSet) One() monoid.CommutativeMonoid {
	n := make(map[string]struct{})
	for k := range f.mine {
		n[k] = struct{}{}
	}
	return &StringSet{
		mine: n,
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
		if _, ok := w.mine[vc] ; ok {
			w.mine[vc]++
		} else {
			w.mine[vc] = 1
		}
	}
	return w
}

func (f WordCount) One() monoid.CommutativeMonoid {
	n := make(map[CopyableString]int)
	for k, v := range f.mine {
		n[k] = v
	}
	return &WordCount{
		mine: n,
	}
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
		if _, ok := n[k] ; ok{
			n[k] = n[k] + v
		} else {
			n[k] = v
		}
	}
	return &WordCount{
		mine: n,
	}
}

func TestWordCount(t *testing.T) {
	Convey("short word count example", t, func() {
		testData := []CopyableString{
			"The The is a great band, yes is also a great one.",
			"the person who wrote the other line has terrible taste",
		}
		
		creator := func(c chan monoid.CommutativeMonoid) {
			out := make(chan interface{}, 100)
			mapping.ForEach(
				func(c chan mapping.Copyable){
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
			"a":2, 
			"also":1, 
			"band":1, 
			"great":2, 
			"has":1, 
			"is":2, 
			"line":1, 
			"one":1, 
			"other":1, 
			"person":1, 
			"taste":1, 
			"terrible":1, 
			"the":4, 
			"who":1, 
			"wrote":1, 
			"yes":1,
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

}
