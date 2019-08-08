package monoid

type CommutativeMonoid interface {
	One() CommutativeMonoid
	Two(a CommutativeMonoid) CommutativeMonoid
}