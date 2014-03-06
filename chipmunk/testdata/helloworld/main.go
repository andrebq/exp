package main

import (
	"fmt"
	cp "github.com/andrebq/exp/chipmunk"
)

func main() {
	gravity := cp.V(0, -100)

	space := cp.NewSpace()
	space.SetGravity(gravity)
	defer space.Free()

	ground := cp.NewSegmentShape(space.StaticBody(), cp.V(-20, 5), cp.V(20, -5), 0)
	ground.SetFriction(1)
	space.AddShape(ground)
	defer ground.Free()

	radius := float32(5.0)
	mass := float32(1.0)

	moment := cp.MomentForCircle(mass, 0, radius, cp.VZero())

	ball := space.AddBody(cp.NewBody(mass, moment))
	ball.SetPos(cp.V(0, 15))
	defer ball.Free()

	space.AddShape(cp.NewCircleShape(ball, radius, cp.VZero()))

	timeStep := float32(1.0 / 60.0)

	// simulate for 2 seconds
	for time := float32(0); time < 2; time += timeStep {
		fmt.Printf("Time: %v / Ball at: %v. Velocity: %v\n",
			time, ball.Pos(), ball.Vel())
		space.Step(timeStep)
	}
}
