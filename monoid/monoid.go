// Copyright 2019 F5 Networks. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package monoid

// Commutative Monoids implement a set of elements with an identity and a communitive associative binary operation.
type CommutativeMonoid interface {
	// binary operator
	Two(a CommutativeMonoid) CommutativeMonoid
}

type Identity func() CommutativeMonoid