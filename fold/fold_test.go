package fold

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/weberr13/go-reducers/monoid"
)

type Foldable struct {
	mine map[string]struct{}
}

func (f Foldable) One() monoid.CommutativeMonoid {
	n := make(map[string]struct{})
	for k := range f.mine {
		n[k] = struct{}{}
	}
	return &Foldable{
		mine: n,
	}
}
func (f Foldable) Two(b monoid.CommutativeMonoid) monoid.CommutativeMonoid {
	realB, ok := b.(*Foldable)
	if !ok {
		panic("Runtime error!  Dynamic cast failure for monoid.CommutativeMonoid -> Foldable")
	}
	n := make(map[string]struct{})
	for k := range f.mine {
		n[k] = struct{}{}
	}
	for k := range realB.mine {
		n[k] = struct{}{}
	}
	return &Foldable{
		mine: n,
	}
}

func TestExampleFoldSlice(t *testing.T) {
	Convey("simple set union fold", t, func() {
		s := []monoid.CommutativeMonoid{
			&Foldable{map[string]struct{}{"foo": struct{}{}}},
			&Foldable{map[string]struct{}{"bar": struct{}{}}},
			&Foldable{map[string]struct{}{"baz": struct{}{}}},
			&Foldable{map[string]struct{}{"foo": struct{}{}}},
		}
		c := FoldSlice(s)
		cr, ok := c.(*Foldable)
		So(ok, ShouldBeTrue)
		So(cr, ShouldResemble, &Foldable{map[string]struct{}{
			"foo": struct{}{},
			"bar": struct{}{},
			"baz": struct{}{},
		}})
	})

}

func TestExampleFoldChan(t *testing.T) {
	Convey("simple set union fold", t, func() {
		s := make(chan monoid.CommutativeMonoid, 10)
		go func () {
			s <- &Foldable{map[string]struct{}{"foo": struct{}{}}}
			s <- &Foldable{map[string]struct{}{"bar": struct{}{}}}
			s <- &Foldable{map[string]struct{}{"baz": struct{}{}}}
			s <- &Foldable{map[string]struct{}{"foo": struct{}{}}}
			close(s)
		}()

		c := FoldChan(s)
		cr, ok := c.(*Foldable)
		So(ok, ShouldBeTrue)
		So(cr, ShouldResemble, &Foldable{map[string]struct{}{
			"foo": struct{}{},
			"bar": struct{}{},
			"baz": struct{}{},
		}})
	})

}
