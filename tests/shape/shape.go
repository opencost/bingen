package shape

import "math"

// Shape is a simple interface implemented by several concrete shapes. It is used
// to exercise interface-typed field encoding/decoding.
type Shape interface {
	Area() float64
}

// Circle is a Shape.
type Circle struct {
	Radius float64
}

// Area implements Shape.
func (c *Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// Square is a Shape.
type Square struct {
	Side float64
}

// Area implements Shape.
func (s *Square) Area() float64 {
	return s.Side * s.Side
}

// Drawing holds a name and a heterogeneous list of shapes.
type Drawing struct {
	Name   string
	Shapes []Shape
}
