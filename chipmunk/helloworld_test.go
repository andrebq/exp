package chipmunk

import (
	"testing"
)

func HelloWorldTest(t *testing.T) {
	t.Logf("This is the helloworld from Chipmunk docs.")
	t.Logf("No output is required, this should just compile and run.")

	gravity := V(0, -100)

	space := NewSpace()
	space.SetGravity(gravity)
	defer space.Free()

	ground := NewSegmentShape(space.staticBody, V(-20, 5), V(20, -5), 0)
	ground.SetFriction(1)
	space.AddShape(ground)
	defer ground.Free()

	radius := float32(5)
	mass := float32(1)

	moment := MomentForCircle(mass, 0, radius, VZero())

	ball := space.AddBody(NewBody(mass, moment))
	ball.SetPos(V(0, 15))
	defer ball.Free()

	ballShape := space.AddShape(NewCircleShape(ball, radius, VZero()))

	timeStep := float32(1.0/60.0)

	// simulate for 2 seconds
	for time := 0; time < 2; time += timeStep {
		t.Logf("Time: %v / Ball at: %v. Velocity: %v",
			time, ball.GetPos(), ball.GetVel())
	}
}
