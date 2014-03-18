package phygo

import (
	glm "github.com/Agon/googlmath"
)

type WorldFlag uint

const (
	NewFixtureFlag = WorldFlag(0x0001)
)

type World struct {
	nextBodyId     Id
	bodyList       bodyList
	ContactManager *ContactManager
	Flags          WorldFlag
}

type bodyList []Body

func (bl *bodyList) alloc() *Body {
	slice := *bl
	if len(slice) == cap(slice) {
		// reached max cap
		tmp := make([]Body, len(slice), len(slice)*2+1)
		copy(tmp, slice)
		slice = tmp
		*bl = slice
	}
	slice = slice[:len(slice)+1]
	return &slice[len(slice)-1]
}

func NewWorld() *World {
	return &World{
		nextBodyId: 1,
		bodyList:   make(bodyList, 0),
		ContactManager: &ContactManager{
			BroadPhase: &BroadPhase{},
		},
	}
}

func (w *World) CreateBody(bd *BodyDef) *Body {
	ret := w.bodyList.alloc()
	BodyFromDef(bd, ret)
	ret.id = w.nextBodyId
	w.nextBodyId++
	return ret
}

func (w *World) CreateJoint(jd *FrictionJointDef) *Joint {
	return nil
}

func (w *World) SetGravity(gravity glm.Vector2) {
}
