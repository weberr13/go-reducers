package monoid

// Commutative Monoids implement a set of elements with an identity and a communitive associative binary operation.
type CommutativeMonoid interface {
	// Identity operation 
	One() CommutativeMonoid
	// binary operator
	Two(a CommutativeMonoid) CommutativeMonoid
}