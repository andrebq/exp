package figo

import (
	"bytes"
	"fmt"
	glm "github.com/Agon/googlmath"
)

type Shape struct {
	body *Body
	poly glm.Polygon
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
	return &Shape{poly: *s, body: nil}
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
	ret.SetScale(s.Scale())
}

func (s *Shape) SetOrigin(origin glm.Vector2) {
	s.poly.SetOrigin(origin)
}

func (s *Shape) Origin() glm.Vector2 {
	return s.poly.Origin()
}

func (s *Shape) SetPosition(pos glm.Vector2) {
	s.poly.SetPosition(pos)
	s.invalidateBody()
}

func (s *Shape) Position() glm.Vector2 {
	return s.poly.Position()
}

func (s *Shape) SetRotation(rot float32) {
	s.poly.SetRotation(rot)
	s.invalidateBody()
}

func (s *Shape) Rotation() float32 {
	return s.poly.Rotation()
}

func (s *Shape) SetScale(scale glm.Vector2) {
	s.poly.SetScale(scale)
	s.invalidateBody()
}

func (s *Shape) Scale() glm.Vector2 {
	return s.poly.Scalar()
}

func (s *Shape) Vertices() []float32 {
	return s.poly.Vertices()
}

func (s *Shape) TransformedVertices() []float32 {
	return s.poly.TransformedVertices()
}

func (s *Shape) AABB() AABB {
	r := s.poly.BoundingRectangle()
	return NewAABB(r.X, r.Y, r.Width, r.Height)
}

func (s *Shape) Translate(transf glm.Vector2) {
	s.poly.Translate(transf)
	s.invalidateBody()
}

func (s *Shape) invalidateBody() {
	if s.body != nil {
		s.body.invalidRect = true
	}
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
