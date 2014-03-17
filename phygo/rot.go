package phygo

import (
	glm "github.com/Agon/googlmath"
)

type Rotation struct {
	Sin, Cos float32
}

func IdentityRotation() Rotation {
	return Rotation{Sin: 0, Cos: 1}
}

func RotationFromAngle(angle float32) Rotation {
	return Rotation{Sin: glm.Sin(angle), Cos: glm.Cos(angle)}
}

func (r *Rotation) MulTo(v glm.Vector2) glm.Vector2 {
	return glm.Vector2{
		r.Cos*v.X - r.Sin*v.Y,
		r.Sin*v.X + r.Cos*v.Y,
	}
}
