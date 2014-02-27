package main

import (
	"github.com/go-gl/gl"
	"github.com/go-gl/glh"
	glfw "github.com/go-gl/glfw3"
	"fmt"
)

var (
	triangleVertex []float32 = []float32{
		-1, -1, 0,
		1, -1, 0,
		0, 1, 0,
	}
	triangleIdx []uint32 = []uint32 {
		0, 1, 2,
	}
)

var theTriangleBuf *glh.MeshBuffer

func updateScene() {
}

func prepareScene() {
	theTriangleBuf = glh.NewMeshBuffer(
		glh.RenderArrays,

		glh.NewIndexAttr(1, gl.UNSIGNED_INT, gl.STATIC_DRAW),
		glh.NewPositionAttr(3, gl.FLOAT, gl.STATIC_DRAW),
	)
	theTriangleBuf.Add(triangleIdx, triangleVertex)
}

func drawScene() {
	theTriangleBuf.Render(gl.TRIANGLES)
}

func errorCallback(err glfw.ErrorCode, desc string) {
	fmt.Printf("%v: %v\n", err, desc)
}

func main() {
	glfw.SetErrorCallback(errorCallback)

	if !glfw.Init() {
		panic("Can't init glfw!")
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(800, 640, "MyGui", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	glfw.SwapInterval(1)
	gl.Init()

	prepareScene()

	for !window.ShouldClose() {
		updateScene()
		drawScene()
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
