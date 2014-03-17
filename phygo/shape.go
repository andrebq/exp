package phygo

import (
	glm "github.com/Agon/googlmath"
)

type Shape struct {
}

func NewEdgeShape() Shape {
	return Shape{}
}

func NewPolygonShape() Shape {
	return Shape{}
}

func (s *Shape) Set(points ...glm.Vector2) {
}

func (s *Shape) SetAsBox(halfSize glm.Vector2) {
}
