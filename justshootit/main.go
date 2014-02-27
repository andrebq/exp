package main

import (
	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	glm "github.com/Agon/googlmath"
	"github.com/go-gl/glh"
	"log"
)

const (
	Title = "Just Shoot-It"
	Width = 800
	Height = 600
)

var (
	timeDelta = &TimeDelta{GetTime: glfw.GetTime, MaxDelta: 0.020}
)

func glfwError(err glfw.ErrorCode, cause string) {
	log.Printf("Error. Code: %V. Cause: %v", err, cause)
}

func initGL() error {
	gl.Init()
	if err := glh.CheckGLError(); err != nil {
		return err
	}

	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.MULTISAMPLE)
	gl.Disable(gl.LIGHTING)
	gl.Enable(gl.COLOR_MATERIAL)
	gl.ClearColor(0, 0, 0, 1)

	return nil
}

// Called to adjust the projection matrix
func fixProjection(win *glfw.Window, w,h int) {
	if w < 1 {
		w = 1
	}

	if h < 1 {
		h = 1
	}

	gl.Viewport(0, 0, w, h)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	// it's a 2d game, so use orthographic projection
	gl.Ortho(0, float64(w), float64(h), 0, 0, 1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

func createMeshBuffer() *glh.MeshBuffer {
	mb := glh.NewMeshBuffer(
		glh.RenderBuffered,

		glh.NewIndexAttr(1, gl.UNSIGNED_SHORT, gl.STATIC_DRAW),
		
		// 2 INTEGER values should be enough, but
		// just to keep things simpler
		// let's work with 3d vertex instead of 2d
		// it's easier to find help that way
		glh.NewPositionAttr(3, gl.FLOAT, gl.STATIC_DRAW),

		// At this moment, let's use colors instead of textures
		glh.NewColorAttr(3, gl.UNSIGNED_BYTE, gl.DYNAMIC_DRAW))
	return mb
}

func vec3ToFloat32(vec *glm.Vec3) *[3]float32 {
	return (*[3]float32)(unsafe.Pointer(&vec.X))
}

// add a new mesh to the mesh buffer that can be used
// to render the square with the given width/height.
//
// the square is centered at it's origion, ie,
// the top-left is width/2, height/2
func addSprite(mb *glh.MeshBuffer, width, height int) {
	vec3 := glm.Vec3(float32(width), float32(height), 0).Nor()

}

func main() {
	glfw.SetErrorCallback(glfwError)

	if !glfw.Init() {
		log.Printf("Unable to initializer glfw")
		return
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(Width, Height, Title, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	window.SetSizeCallback(fixProjection)

	err = initGL()
	if err != nil {
		log.Printf("Error initializing OpenGL. %v", err)
		panic(err)
	}

	glfw.SwapInterval(1)
	fixProjection(window, Width, Height)

	meshBuff := createMeshBuffer()

	for !window.ShouldClose() {
		timeDelta.Tick()
		log.Printf("Time: %v", timeDelta.Delta)
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

