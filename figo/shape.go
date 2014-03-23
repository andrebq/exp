package figo

import (
	"bytes"
	"fmt"
	glm "github.com/Agon/googlmath"
)

type Shape struct {
	*glm.Polygon
	body *Body
}

// NewRectShape create a new Shape and
// In this function you should provide half of the
// final size, if you give (1,1) you get a box of
// (2,2).
//
// X and Y don't need to be equal
//
// Vertex information is set clockwise.
func NewRectShape(halfSize glm.Vector2) *Shape {
	var points [8]float32
	points[0], points[1] = -halfSize.X, halfSize.Y
	points[2], points[3] = halfSize.X, halfSize.Y
	points[4], points[5] = halfSize.X, -halfSize.Y
	points[6], points[7] = -halfSize.X, -halfSize.Y
	return polygonShapeFromFloatSlice(points[:]...)
}

// NewPolygonShape will copy all vertices from points
// to this shape.
func NewPolygonShape(points ...glm.Vector2) *Shape {
	return polygonShapeFromFloatSlice(VecToFloat(nil, points...)...)
}

func polygonShapeFromFloatSlice(points ...float32) *Shape {
	s, _ := glm.NewPolygon(points)
	return &Shape{s, nil}
}

func (s *Shape) Clone() *Shape {
	ret := polygonShapeFromFloatSlice(s.Vertices()...)
	s.applyChangesTo(ret)
	return ret
}

// copy all local modifications from s to ret
func (s *Shape) applyChangesTo(ret *Shape) {
	ret.SetOrigin(s.Origin())
	ret.SetPosition(s.Position())
	ret.SetRotation(s.Rotation())
	ret.SetScale(s.Scalar())
}

func (s *Shape) String() string {
	buf := &bytes.Buffer{}
	vert := s.Vertices()
	fmt.Fprintf(buf, "shape[%v] ", len(vert)/2)
	for i := 0; i < len(vert); i += 2 {
		fmt.Fprintf(buf, "(%v, %v) ", vert[i], vert[i+1])
	}
	return string(buf.Bytes())
}
