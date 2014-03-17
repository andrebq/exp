package phygo

import (
	glm "github.com/Agon/googlmath"
)

type FrictionJointDef struct {
	LocalAnchorA     glm.Vector2
	LocalAnchorB     glm.Vector2
	BodyA            *Body
	BodyB            *Body
	CollideConnected bool
	MaxTorque        float32
	MaxForce         float32
}

func NewFrictionJointDef() FrictionJointDef {
	return FrictionJointDef{}
}
