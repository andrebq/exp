package chipmunk

import (
	"testing"
)

func TestHelloWorld(t *testing.T) {
	t.Logf("This is the helloworld from Chipmunk docs.")
	t.Logf("No output is required, this should just compile and run.")

	gravity := V(0, -100)

	space := NewSpace()
	space.SetGravity(gravity)
	defer space.Free()

	ground := NewSegmentShape(space.StaticBody(), V(-20, 5), V(20, -5), 0)
	ground.SetFriction(1)
	space.AddShape(ground)
	defer ground.Free()

	radius := float32(5)
	mass := float32(1)

	moment := MomentForCircle(mass, 0, radius, VZero())

	ball := space.AddBody(NewBody(mass, moment))
	ball.SetPos(V(0, 15))
	defer ball.Free()

	space.AddShape(NewCircleShape(ball, radius, VZero()))

	timeStep := float32(1.0 / 60.0)

	iPos, iVel := ball.Pos(), ball.Vel()

	// simulate for 2 seconds
	for time := float32(0); time < 2; time += timeStep {
		t.Logf("Time: %v / Ball at: %v. Velocity: %v",
			time, ball.Pos(), ball.Vel())
		space.Step(timeStep)
	}

	if iPos == ball.Pos() {
		t.Errorf("Position should have changed. Initial is: %v actual is: %v",
			iPos, ball.Pos())
	}

	if iVel == ball.Vel() {
		t.Errorf("Velocity should have changed. Initial is: %v actual is: %v",
			iVel, ball.Vel())
	}
}
