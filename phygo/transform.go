package phygo

import (
	glm "github.com/Agon/googlmath"
)

type Transform struct {
	P glm.Vector2
	Q Rotation
}

func NewTransform() Transform {
	return Transform{}
}

func (t *Transform) MulTo(v glm.Vector2) glm.Vector2 {
	return t.Q.MulTo(v).Add(t.P)
}
