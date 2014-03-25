package opala

import (
	glm "github.com/Agon/googlmath"
	"github.com/go-gl/gl"
	"image"
	"image/color"
)

// Texture represents the image that should be rendered
// at a given location.
//
// Textures are applied to a object and combined with a tint
// to give the final output
type Texture struct {
	Image *image.RGBA
	UV    glm.Vector2

	glTextureId gl.Texture
}

// Object is a object that can be renderd at a give location
type Object struct {
	// The texture to apply for this object
	Texture *Texture
	// A color to paint over the texture
	Tint color.Color

	// The center of this object
	center glm.Vector3
	// The scale of this object
	scale glm.Vector3
	// The rotation in degrees for this object
	rotationDeg float32

	// model matrix for vertex shader
	modelMatrix *glm.Matrix4
	// indicates that the object matrix must be updated
	dirty bool
}
