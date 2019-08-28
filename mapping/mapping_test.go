// Copyright 2019 F5 Networks. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mapping_test

import (
	"testing"
	"sync"
	"strings"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/weberr13/go-reducers/mapping"
)

type CopyableStringSlice struct {
	mine []string
}

func (c CopyableStringSlice) Copy() mapping.Copyable {
	n := make([]string, len(c.mine))
	copy(n, c.mine)
	return &CopyableStringSlice{
		mine: n,
	}
}

func TestForEach(t *testing.T) {
	tests := map[string]struct{
		s []CopyableStringSlice
		n int
		f func(string) string
	}{
		"For Each String": {
			s: []CopyableStringSlice{
				CopyableStringSlice{mine: []string{"a", "B", "C"}},
				CopyableStringSlice{mine: []string{"d", "E", "F"}},
			},
			n: 32,
			f: strings.ToLower,
		},
		"For Each String, lower thread boundary": {
			s: []CopyableStringSlice{
				CopyableStringSlice{mine: []string{"a", "B", "C"}},
				CopyableStringSlice{mine: []string{"d", "E", "F"}},
			},
			n: -1,
			f: strings.ToLower,
		},
		"For Each String, upper thread boundary": {
			s: []CopyableStringSlice{
				CopyableStringSlice{mine: []string{"a", "B", "C"}},
				CopyableStringSlice{mine: []string{"d", "E", "F"}},
			},
			n: 1025,
			f: strings.ToLower,
		},
	}
	for name, ts := range tests {
		Convey(name, t, func() {
			wg := &sync.WaitGroup{}
			out := make(chan interface{}, 1024)
			wg.Add(1)

			go func() {
				defer wg.Done()
				mapping.ForEachN(
					func(c chan<- mapping.Copyable) {
						for _, v := range ts.s {
							c <- v
						}
						close(c)
					},
					out,
					func(v mapping.Copyable) interface{} {
						vc := v.(*CopyableStringSlice)
						for i := range vc.mine {
							vc.mine[i] = ts.f(vc.mine[i])
						}
						return vc
					},
					ts.n,
				)
			}()
			for raw := range out {
				s, ok := raw.(*CopyableStringSlice)
				So(ok, ShouldBeTrue)
				for _, v := range s.mine {
					So(v, ShouldEqual, ts.f(v))
				}
			}
			wg.Wait()
		})
	}
}
