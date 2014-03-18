package phygo

import (
	glm "github.com/Agon/googlmath"
	"testing"
)

func TestApplyForce(t *testing.T) {
	world := getTestWorld()
	world.SetGravity(glm.Vector2{0, 0})

	// creating body
	def := NewBodyDef()
	def.Position = def.Position.Set(0, 20)

	// restitution
	Restitution := float32(0.4)

	ground := world.CreateBody(&def)

	shape := NewEdgeShape()
	fixtureDef := NewFixtureDef()
	fixtureDef.Shape = shape
	fixtureDef.Density = 0.0
	fixtureDef.Restitution = Restitution

	// left shape
	shape.Set(glm.Vector2{-20, -20}, glm.Vector2{-20, 20})
	if ground.CreateFixture(&fixtureDef, world) == nil {
		t.Fatalf("unable to create fixture for body: %v with def: %v",
			ground, fixtureDef)
	}

	// right shape
	shape.Set(glm.Vector2{20, -20}, glm.Vector2{20, 20})
	if ground.CreateFixture(&fixtureDef, world) == nil {
		t.Fatalf("unable to create fixture for body: %v with def: %v",
			ground, fixtureDef)
	}

	// top
	shape.Set(glm.Vector2{-20, 20}, glm.Vector2{20, 20})
	if ground.CreateFixture(&fixtureDef, world) == nil {
		t.Fatalf("unable to create fixture for body: %v with def: %v",
			ground, fixtureDef)
	}

	// bottom
	shape.Set(glm.Vector2{-20, -20}, glm.Vector2{20, -20})
	if ground.CreateFixture(&fixtureDef, world) == nil {
		t.Fatalf("unable to create fixture for body: %v with def: %v",
			ground, fixtureDef)
	}

	xf1 := NewTransform()
	xf1.Q = RotationFromAngle(0.3524 * glm.Pi)
	xf1.P = xf1.Q.MulTo(glm.Vector2{1, 0})

	vertices := make([]glm.Vector2, 3)
	vertices[0] = xf1.MulTo(glm.Vector2{-1, 0})
	vertices[1] = xf1.MulTo(glm.Vector2{1, 0})
	vertices[2] = xf1.MulTo(glm.Vector2{0, 0.5})

	poly1 := NewPolygonShape()
	poly1.Set(vertices...)

	sd1 := NewFixtureDef()
	sd1.Shape = poly1
	sd1.Density = 4

	xf2 := NewTransform()
	xf2.Q = RotationFromAngle(-0.3524 * glm.Pi)
	xf2.P = xf2.Q.MulTo(glm.Vector2{-1, 0})

	vertices[0] = xf2.MulTo(glm.Vector2{-1, 0})
	vertices[1] = xf2.MulTo(glm.Vector2{1, 0})
	vertices[2] = xf2.MulTo(glm.Vector2{0, 0.5})

	poly2 := NewPolygonShape()
	poly2.Set(vertices...)

	sd2 := NewFixtureDef()
	sd2.Shape = poly2
	sd2.Density = 2

	bd := NewBodyDef()
	bd.Type = DYNAMIC
	bd.AngularDamping = 2
	bd.LinearDamping = 0.5
	bd.Position = glm.Vector2{0, 2}
	bd.Angle = glm.Pi
	bd.AllowSleep = false

	m_body := world.CreateBody(&bd)
	if m_body == nil {
		t.Fatalf("unable to create body with def: %v", bd)
	}
	m_body.CreateFixture(&sd1, world)
	m_body.CreateFixture(&sd2, world)

	shape = NewPolygonShape()
	shape.SetAsBox(glm.Vector2{0.5, 0.5})

	fd := NewFixtureDef()
	fd.Shape = shape
	fd.Density = 1
	fd.Friction = 0.3

	for i := 0; i < 10; i++ {
		bd := NewBodyDef()
		bd.Type = DYNAMIC
		bd.Position = glm.Vector2{0, 5.0 + 1.54*float32(i)}

		body := world.CreateBody(&bd)
		body.CreateFixture(&fd, world)

		gravity := float32(10.0)
		inertia := body.Inertia()
		mass := body.Mass()
		radius := glm.Sqrt(2.0 * inertia / mass)

		jd := NewFrictionJointDef()
		jd.LocalAnchorA = glm.Vector2{0, 0}
		jd.LocalAnchorB = glm.Vector2{0, 0}
		jd.BodyA = ground
		jd.BodyB = body
		jd.CollideConnected = true
		jd.MaxForce = mass * gravity
		jd.MaxTorque = mass * radius * gravity

		if world.CreateJoint(&jd) == nil {
			t.Fatalf("unable to create joint using def: %v", jd)
		}
	}
}

func getTestWorld() *World {
	return NewWorld()
}
