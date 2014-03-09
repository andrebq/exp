package chipmunk

import (
	"testing"
)

func PyramidStackSpace(callback func(s *Space)) {
	grab_enable := Layer(1 << 31)
	grab_disable := grab_enable.Not()

	space := NewSpace()
	space.SetIterations(30)
	space.SetGravity(V(0, -100))
	space.SetSleepTimeThreshold(0.5)
	space.SetCollisionSlop(0.5)

	staticBody := space.StaticBody()

	// creating the edges
	shape1 := space.AddShape(NewSegmentShape(staticBody, V(-320, -240), V(-320, 240), 0))
	shape1.SetElasticity(1.0)
	shape1.SetFriction(1.0)
	shape1.SetLayers(grab_disable)
	defer shape1.Free()

	shape2 := space.AddShape(NewSegmentShape(staticBody, V(320, -240), V(320, -240), 0))
	shape2.SetElasticity(1.0)
	shape2.SetFriction(1.0)
	shape2.SetLayers(grab_disable)
	defer shape2.Free()

	shape3 := space.AddShape(NewSegmentShape(staticBody, V(-320, -240), V(320, -240), 0))
	shape3.SetElasticity(1.0)
	shape3.SetFriction(1.0)
	shape3.SetLayers(grab_disable)
	defer shape3.Free()

	for i := 0; i < 14; i++ {
		for j := 0; j < i; j++ {
			body := space.AddBody(NewBody(1, MomentForBox(1, 30, 30)))
			body.SetPos(V(float32(j)*32-float32(i)*16, 300-float32(i)*32))

			shape := space.AddShape(NewBoxShape(body, 30, 30))
			shape.SetElasticity(0)
			shape.SetFriction(0.8)
			defer shape.Free()
			defer body.Free()
		}
	}

	radius := float32(15)
	body := space.AddBody(NewBody(10, MomentForCircle(10, 0, radius, VZero())))
	body.SetPos(V(0, -240+radius+5))

	shape4 := space.AddShape(NewCircleShape(body, radius, VZero()))
	shape4.SetElasticity(0)
	shape4.SetFriction(0.9)
	defer shape4.Free()

	defer body.Free()
	defer space.Free()

	callback(&space)
}

// Demo from Chipmunk/Demo/PyramidStack.c
func TestPyramidStack(t *testing.T) {
	PyramidStackSpace(func(s *Space) {
		s.StepSeconds(2, 1/60.0)
	})
}

func BenchmarkPyramidStack(b *testing.B) {
	PyramidStackSpace(func(s *Space) {
		for i := 0; i < b.N; i++ {
			s.Step(1 / 60.0)
		}
	})
}
