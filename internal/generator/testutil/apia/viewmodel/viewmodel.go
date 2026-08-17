// Package viewmodel is one half of a pair of same-named packages declaring the same
// type name, the shape that makes a short annotation name ambiguous to swag.
package viewmodel

type LoginResponse struct {
	TokenA string
}

type OnlyInA struct {
	Name string
}
