package chipmunk

import (
	"testing"
)

// Demo from Chipmunk/Demo/PyramidStack.c
func TestPyramidStack(t *testing.T) {
	grab_enable := Layer(1 << 31)
	grab_disable := grab_enable.Not()
	tofree := newFreeStack()

	space := NewSpace()
	space.SetIterations(30)
	space.SetGravity(V(0, -100))
	space.SetSleepTimeThreshold(0.5)
	space.SetCollisionSlop(0.5)
	tofree.push(&space)

	staticBody := space.StaticBody()

	// creating the edges
	shape := space.AddShape(NewSegmentShape(staticBody, V(-320, -240), V(-320, 240), 0))
	shape.SetElasticity(1.0)
	shape.SetFriction(1.0)
	shape.SetLayers(grab_disable)
}
