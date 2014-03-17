package phygo

type FixtureDef struct {
	Shape       Shape
	Density     float32
	Restitution float32
	Friction    float32
}

func NewFixtureDef() FixtureDef {
	return FixtureDef{}
}
