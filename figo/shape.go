package figo

import (
	glm "github.com/Agon/googlmath"
)

type Shape struct {
	*glm.Polygon
}

// NewRectShape create a new Shape and
// In this function you should provide half of the
// final size, if you give (1,1) you get a box of
// (2,2).
//
// X and Y don't need to be equal
//
// Vertex information is set clockwise.
func NewRectShape(halfSize glm.Vector2) Shape {
	var points [8]float32
	points[0], points[1] = -halfSize.X, halfSize.Y
	points[2], points[3] = halfSize.X, halfSize.Y
	points[4], points[5] = halfSize.X, -halfSize.Y
	points[6], points[7] = -halfSize.X, -halfSize.Y
	s, _ := glm.NewPolygon(points[:])
	return Shape{s}
}

// SetAsPolygon will copy all vertices from points
// to this shape.
func SetAsPolygon(points ...glm.Vector2) Shape {
	s, _ := glm.NewPolygon(VecToFloat(nil, points...))
	return Shape{s}
}
