package monoid

type CommunativeMonoid interface {
	One() CommunativeMonoid
	Two(a CommunativeMonoid) CommunativeMonoid
}