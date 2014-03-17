package phygo

import (
	glm "github.com/Agon/googlmath"
)

type Transform struct {
	P glm.Vector2
	Q Rotation
}

func NewTransform() Transform {
	return Transform{
		Q: IdentityRotation(),
	}
}

func NewTransformFromPosAngle(pos glm.Vector2, angle float32) Transform {
	return Transform{
		P: pos,
		Q: RotationFromAngle(angle),
	}
}

func (t *Transform) MulTo(v glm.Vector2) glm.Vector2 {
	return t.Q.MulTo(v).Add(t.P)
}
