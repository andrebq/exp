package main

import (
	"fmt"
	glm "github.com/Agon/googlmath"
	"github.com/andrebq/gas"
	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/glh"
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
	triangleIdx []uint32 = []uint32{
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
	vertexShader, err := compileVertexShader(filepath.Join(dir, program+".vertex"))
	if err != nil {
		return
	}
	defer vertexShader.Delete()
	fragShader, err := compileFragmentShader(filepath.Join(dir, program+".fragment"))
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
	shaderDir      = gas.MustAbs("github.com/andrebq/exp/gui")
)

func updateScene(dt float64) {
}

func prepareScene() {
	theTriangleBuf = glh.NewMeshBuffer(
		glh.RenderArrays,

		glh.NewIndexAttr(1, gl.UNSIGNED_INT, gl.STATIC_DRAW),
		glh.NewPositionAttr(3, gl.FLOAT, gl.STATIC_DRAW),
	)
	theTriangleBuf.Add(triangleIdx, triangleVertex)
}

var (
	triangleScale             = float32(10)
	triangleScaleChangeFactor = float32(200)
)

func drawScene(mvp *glm.Matrix4, dt float64) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	program, err := compileProgram(shaderDir, "sample")
	if err != nil {
		log.Printf("Error reading program. Cause: %v", err)
	} else {
		defer program.Delete()
	}

	// lets scale our model by 100, otherwise
	// the triangle will be 1 pixel width
	// let's make our triangle pulsate
	triangleScale += triangleScaleChangeFactor * float32(dt)
	if triangleScale >= 200 {
		triangleScale = 200
		triangleScaleChangeFactor *= -1
	} else if triangleScale < 10 {
		triangleScale = 10
		triangleScaleChangeFactor *= -1
	}
	scaled := mvp.Scale(glm.Vector3{triangleScale, triangleScale, triangleScale})

	loc := program.GetUniformLocation("MVP")
	loc.UniformMatrix4f(false, ptrForMatrix(scaled))
	theTriangleBuf.Render(gl.TRIANGLES)
}

func ptrForMatrix(m *glm.Matrix4) *[16]float32 {
	ret := [16]float32{
		m.M11, m.M12, m.M13, m.M14,
		m.M21, m.M22, m.M23, m.M24,
		m.M31, m.M32, m.M33, m.M34,
		m.M41, m.M42, m.M43, m.M44,
	}
	return &ret
}

func errorCallback(err glfw.ErrorCode, desc string) {
	fmt.Printf("%v: %v\n", err, desc)
}

func prepareModelViewProjection(width, height float32) *glm.Matrix4 {
	modelM := glm.NewIdentityMatrix4()
	// this is just 2d, so ignore the camera matrix
	// camera at 0,0,1 (could be anything > 0 )
	// looking at 0,0,0
	// with the +Y is the Up vector
	//
	// could be replaced by a Identity matrix, but keeping this here
	// since it make easier to turn the world upside down
	viewM := glm.NewLookAtMatrix4(
		glm.Vector3{0, 0, 1},
		glm.Vector3{0, 0, 0},
		glm.Vector3{0, -1, 0})

	// in order to keep things simple,
	// the 2d projection uses the center of screen as the point 0,0
	// ie, the (top,left) == (-width/2, heiht/2)
	projection := glm.NewOrthoMatrix4(-width/2, width/2, -height/2, height/2, 0, 100)

	return projection.Mul(viewM).Mul(modelM)
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

	mvp := prepareModelViewProjection(800, 640)

	prepareScene()
	dt := float64(0)
	glfw.SetTime(dt)

	for !window.ShouldClose() {
		dt = glfw.GetTime()
		glfw.SetTime(0)
		updateScene(dt)
		drawScene(mvp, dt)
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
