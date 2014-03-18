package phygo

type FixtureDef struct {
	Shape       *Shape
	Density     float32
	Restitution float32
	Friction    float32
	Filter      Filter
	UserData    uint64
	IsSensor    bool
}

func NewFixtureDef() FixtureDef {
	return FixtureDef{}
}
