package phygo

import (
	glm "github.com/Agon/googlmath"
)

type World struct {
}

func NewWorld() *World {
	return &World{}
}

func (w *World) CreateBody(bd *BodyDef) *Body {
	return nil
}

func (w *World) CreateJoint(jd *FrictionJointDef) *Joint {
	return nil
}

func (w *World) SetGravity(gravity glm.Vector2) {
}
