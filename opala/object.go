package opala

import (
	glm "github.com/Agon/googlmath"
	"image/color"
)

// Object is a object that can be renderd at a give location
type Object struct {
	// The piece of the atlas that should
	// be used to render this object
	Atlas *AtlasChunk
	// model matrix for vertex shader
	modelMatrix *glm.Matrix4
	// A color to paint over the texture
	Tint color.Color

	// The center of this object
	center glm.Vector3
	// The scale of this object
	scale glm.Vector3
	// The rotation in degrees for this object
	rotationDeg float32

	// indicates that the object matrix must be updated
	dirty bool
}

func (o *Object) render(dt float32) {
}
