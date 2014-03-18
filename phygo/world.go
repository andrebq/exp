package phygo

import (
	glm "github.com/Agon/googlmath"
)

type WorldFlag uint

const (
	NewFixtureFlag = WorldFlag(0x0001)
)

type World struct {
	rootBody       *Body
	nextBodyId     Id
	ContactManager *ContactManager
	Flags          WorldFlag
}

func NewWorld() *World {
	return &World{
		nextBodyId: 1,
		ContactManager: &ContactManager{
			BroadPhase: &BroadPhase{},
		},
	}
}

func (w *World) CreateBody(bd *BodyDef) *Body {
	body := &Body{}
	BodyFromDef(bd, body, w)
	if w.rootBody == nil {
		w.rootBody = body
	} else {
		body.next = w.rootBody
		w.rootBody.prev = body
		w.rootBody = body
	}
	body.id = w.nextBodyId
	w.nextBodyId++
	return body
}

func (w *World) CreateJoint(jd *FrictionJointDef) *Joint {
	return nil
}

func (w *World) SetGravity(gravity glm.Vector2) {
}
