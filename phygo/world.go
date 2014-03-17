package phygo

import (
	glm "github.com/Agon/googlmath"
)

type World struct {
	nextBodyId Id
	bodyList   []Body
}

func NewWorld() *World {
	return &World{
		nextBodyId: 1,
		bodyList:   make([]Body, 0, 10000),
	}
}

func (w *World) CreateBody(bd *BodyDef) *Body {
	w.bodyList = append(w.bodyList, NewBody(bd))
	ret := &w.bodyList[len(w.bodyList)-1]
	ret.id = w.nextBodyId
	w.nextBodyId++
	return ret
}

func (w *World) CreateJoint(jd *FrictionJointDef) *Joint {
	return nil
}

func (w *World) SetGravity(gravity glm.Vector2) {
}
