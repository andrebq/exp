package phygo

import (
	"fmt"
	glm "github.com/Agon/googlmath"
)

type BodyFlag uint

const (
	AutoSleepFlag = BodyFlag(0x0004)
)

type Body struct {
	id    Id
	Flags BodyFlag
	// body origin transform
	Xf Transform
	// previous transform for particle simulation
	Xf0             Transform
	Type            BodyType
	IslandIndex     int
	LinearVelocity  glm.Vector2
	AngularVelocity float32
	Force           glm.Vector2
	Torque          float32
	Fixture         []Fixture
	SleepTime       float32

	LinearDamping  float32
	AngularDamping float32
	GravityScale   float32

	mass       float32
	invMass    float32
	inertia    float32
	invInertia float32
}

func (b *Body) String() string {
	return fmt.Sprintf("body[%v]", b.id)
}

func NewBody(bd *BodyDef) Body {
	ret := Body{
		Xf:              NewTransformFromPosAngle(bd.Position, bd.Angle),
		Xf0:             NewTransform(),
		LinearVelocity:  bd.LinearVelocity,
		AngularVelocity: bd.AngularVelocity,
		Type:            bd.Type,
	}

	ret.Flags |= AutoSleepFlag

	if ret.Type == DYNAMIC {
		ret.mass = 1
		ret.invMass = 1
	} else {
		ret.mass = 0
		ret.invMass = 0
	}
	ret.inertia = 0
	ret.invInertia = 0
	ret.Fixture = make([]Fixture, 0, 1)
	return ret
}

func (b *Body) CreateFixture(fd FixtureDef) *Fixture {
	return nil
}

func (b *Body) Inertia() float32 {
	return 0
}

func (b *Body) Mass() float32 {
	return 0
}

// Usually the Id is unique inside a world, the world object
// is the one responsible for given the body it's Id. A body
// without a world will have the invalid id of 0
func (b *Body) Id() Id {
	return b.id
}
