package phygo

type Body struct {
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
