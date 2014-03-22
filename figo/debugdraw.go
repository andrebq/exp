package figo

// Hold information to identify the given frame
type FrameData struct {
	// The fixed delta used in the engine
	FixedDelta float32
	// The time elapsed from the last frame
	Delta float32
	// The number of this frame
	Number uint

	// Half of the world width
	HalfWidth float32

	// Half of the world height
	HalfHeight float32
}

// DebugDraw group a small number of functions used
// by the figo engine to give a visual representation
// of the world at the given state
type DebugDraw interface {
	// BeginFrame is called to signal the start of a new frame
	BeginFrame(frame FrameData)
	// EndFrame is called to signal the end of the last frame
	EndFrame(frame FrameData)
	// DrawShape should the given shape
	//
	// If velocity != 0 the shape is moving
	// If accel != 0 the shape is moving and accelerating
	// If mass != 0 the shape is solid
	DrawShape(s *Shape, mass, velocity, accel float32)
}
