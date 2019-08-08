package fold

import (
	"testing"

	"github.com/weberr13/go-reducers/monoid"
	. "github.com/smartystreets/goconvey/convey"
)

type Foldable struct {
	mine map[string]struct{}
}
func (f Foldable) One() monoid.CommunativeMonoid {
	n := make(map[string]struct{})
	for k := range f.mine {
		n[k] = struct{}{}
	}
	return &Foldable{
		mine: n,
	}
}
func (f Foldable) Two(b monoid.CommunativeMonoid) monoid.CommunativeMonoid {
	realB, ok := b.(*Foldable)
	if !ok {
		panic("Runtime error!  Dynamic cast failure for monoid.CommunativeMonoid -> Foldable")
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
	Convey("simple fold", t, func() {
		s := []monoid.CommunativeMonoid{
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