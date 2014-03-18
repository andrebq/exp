package phygo

type Fixture struct {
	Density float32
	Next    *Fixture
}

type proxyList []FixtureProxy

func (pl *proxyList) alloc() *FixtureProxy {
	slice := *pl
	if len(slice) == cap(slice) {
		tmp := make([]FixtureProxy, len(slice), len(slice)*2+1)
		copy(tmp, slice)
		slice = tmp
		*pl = slice
	}
	slice = slice[:len(slice)+1]
	return &slice[len(slice)-1]
}

func FixtureFromDef(def *FixtureDef, fix *Fixture) {
	fix.Friction = def.Friction
	fix.Restitution = def.Restitution
	fix.UserData = def.UserData
	fix.IsSensor = def.IsSensor
	fix.Shape = def.Shape.Clone()
	fix.Filter = def.Filter
	fix.Next = nil

	childCount := fix.Shape.ChildCount()
	fix.proxyList.reset()
	for i := 0; i < childCount; i++ {
		proxy := fix.proxyList.alloc()
		proxy.fixture = nil
		proxy.proxyId = NullProxy
	}
}

func (f *Fixture) CreateProxies(bp *BroadPhase, bodyTrans *Transform) {
	println("TODO implement this later. i am too lazy to do it now")
}
