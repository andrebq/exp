package phygo

import (
	"fmt"
	glm "github.com/Agon/googlmath"
)

type BodyType byte

const (
	DYNAMIC = BodyType(1)
)

func (bt BodyType) String() string {
	if name, has := bodyTypeName[bt]; has {
		return name
	}
	return fmt.Sprintf("unknown bodytype: %v", bt)
}

var (
	bodyTypeName map[BodyType]string = map[BodyType]string{
		DYNAMIC: "Dynamic",
	}
)

type BodyDef struct {
	Position        glm.Vector2
	Type            BodyType
	AngularDamping  float32
	LinearDamping   float32
	Angle           float32
	AllowSleep      bool
	LinearVelocity  glm.Vector2
	AngularVelocity float32
}

func NewBodyDef() BodyDef {
	return BodyDef{}
}
