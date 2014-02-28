package main

import (
	"github.com/andrebq/gas"
	"github.com/go-gl/gl"
	"github.com/go-gl/glh"
	glfw "github.com/go-gl/glfw3"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
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

func compileFragmentShader(fileName string) (vshader gl.Shader, err error) {
	vshader = gl.CreateShader(gl.FRAGMENT_SHADER)
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	vshader.Source(string(buf))
	vshader.Compile()
	if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		return vshader, fmt.Errorf("Unable to compile fragment shader. Cause: %v", vshader.GetInfoLog())
	}
	return
}

func compileVertexShader(fileName string) (vshader gl.Shader, err error) {
	vshader = gl.CreateShader(gl.VERTEX_SHADER)
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	vshader.Source(string(buf))
	vshader.Compile()
	if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		return vshader, fmt.Errorf("Unable to compile vertex shader. Cause: %v", vshader.GetInfoLog())
	}
	return
}

func compileProgram(dir, program string) (glProg gl.Program, err error) {
	vertexShader, err := compileVertexShader(filepath.Join(dir, program + ".vertex"))
	if err != nil {
		return
	}
	defer vertexShader.Delete()
	fragShader, err := compileFragmentShader(filepath.Join(dir, program + ".fragment"))
	if err != nil {
		return
	}
	defer fragShader.Delete()

	glProg = gl.CreateProgram()
	glProg.AttachShader(vertexShader)
	glProg.AttachShader(fragShader)

	glProg.Link()

	if glProg.Get(gl.LINK_STATUS) != gl.TRUE {
		defer vertexShader.Delete()
		defer fragShader.Delete()
		return glProg, fmt.Errorf("Unable to link glProg. Cause: %v", glProg.GetInfoLog())
	}
	glProg.Use()

	return
}

var (
	theTriangleBuf *glh.MeshBuffer
	shaderDir = gas.MustAbs("github.com/andrebq/exp/gui")
)

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
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	program, err := compileProgram(shaderDir, "sample")
	if err != nil {
		log.Printf("Error reading program. Cause: %v", err)
	} else {
		defer program.Delete()
	}
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
