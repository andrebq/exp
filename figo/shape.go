package figo

import (
	glm "github.com/Agon/googlmath"
)

type Shape struct {
	Vertex []glm.Vector2
}

// SetAsRect update the Shape vertex info,
// and make it a box centered at (0,0)
//
// In this function you should provide half of the
// final size, if you give (1,1) you get a box of
// (2,2).
//
// X and Y don't need to be equal
//
// Vertex information is set clockwise.
func (s *Shape) SetAsRect(halfSize glm.Vector2) {
	s.expandVertexToFit(4)
	s.Vertex[0] = glm.Vector2{-halfSize.X, halfSize.Y}
	s.Vertex[1] = glm.Vector2{halfSize.X, halfSize.Y}
	s.Vertex[2] = glm.Vector2{halfSize.X, -halfSize.Y}
	s.Vertex[3] = glm.Vector2{-halfSize.X, -halfSize.Y}
}

// SetAsPolygon will copy all vertices from points
// to this shape.
func (s *Shape) SetAsPolygon(points ...glm.Vector2) {
	s.expandVertexToFit(len(points))
	copy(s.Vertex, points)
}

// expand the vertex attribute, no data is zeroed
func (s *Shape) expandVertexToFit(newSize int) {
	if cap(s.Vertex) >= newSize {
		s.Vertex = s.Vertex[:newSize]
	} else {
		s.Vertex = make([]glm.Vector2, newSize)
	}
}
