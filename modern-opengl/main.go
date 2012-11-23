// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This program demonstrates the use of a MeshBuffer.
package main

import (
	"github.com/go-gl/gl"
	"github.com/go-gl/glfw"
	"github.com/go-gl/glh"
	"github.com/go-gl/glu"
	"github.com/andrebq/assimp"
	"github.com/andrebq/assimp/conv"
	"github.com/andrebq/gas"
	"log"
)

var (
	scene *assimp.Scene
)

func main() {
	
	loadMeshInfo()
	
	err := initGL()
	if err != nil {
		log.Printf("InitGL: %v", err)
		return
	}

	program := createSampleProgram()
	_ = program

	defer glfw.Terminate()

	mb := createBuffer()
	defer mb.Release()

	// Perform the rendering.
	var angle float32
	for glfw.WindowParam(glfw.Opened) > 0 {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.LoadIdentity()
		gl.Translatef(0, 0, -20)
		gl.Rotatef(angle, 1, 1, 1)
		//program.Use()

		// Render a solid cube at half the scale.
		//gl.Scalef(0.2, 0.2, 0.2)
		gl.Enable(gl.COLOR_MATERIAL)
		gl.Enable(gl.POLYGON_OFFSET_FILL)
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		mb.Render(gl.TRIANGLES)
		/*

		// Render wireframe cubes, with incremental size.
		gl.Disable(gl.COLOR_MATERIAL)
		gl.Disable(gl.POLYGON_OFFSET_FILL)
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

		for i := 0; i < 50; i++ {
			scale := 0.004*float32(i) + 1.0
			gl.Scalef(scale, scale, scale)
			mb.Render(gl.QUADS)
		}
		*/

		angle += 0.5
		glfw.SwapBuffers()
	}
}

func createBuffer() *glh.MeshBuffer {
	
	assimp.RandomColor(scene.Mesh[0])
	fmesh := assimp.NewFlatMesh(scene.Mesh[0])

	// Create a mesh buffer with the given attributes.
	mb := glh.NewMeshBuffer(
		glh.RenderBuffered,

		// Indices.
		glh.NewIndexAttr(1, gl.UNSIGNED_BYTE, gl.STATIC_DRAW),

		// Vertex positions have 3 components (x, y, z).
		glh.NewPositionAttr(3, gl.FLOAT, gl.STATIC_DRAW),

		// Colors have 4 components (r, g, b, a).
		glh.NewColorAttr(4, gl.FLOAT, gl.STATIC_DRAW),
	)

	// Add the mesh to the buffer.
	mb.Add(fmesh.ByteIndex, fmesh.Vertex, fmesh.Color)
	return mb
}

// initGL initializes GLFW and OpenGL.
func initGL() error {
	err := glfw.Init()
	if err != nil {
		return err
	}

	glfw.OpenWindowHint(glfw.FsaaSamples, 4)

	err = glfw.OpenWindow(512, 512, 8, 8, 8, 8, 32, 0, glfw.Windowed)
	if err != nil {
		glfw.Terminate()
		return err
	}

	glfw.SetWindowTitle("Meshbuffer 3D example")
	glfw.SetSwapInterval(1)
	glfw.SetWindowSizeCallback(onResize)
	glfw.SetKeyCallback(onKey)

	gl.Init()
	if err = glh.CheckGLError(); err != nil {
		return err
	}

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.MULTISAMPLE)
	gl.Disable(gl.LIGHTING)

	gl.ClearColor(0.2, 0.2, 0.23, 1.0)
	gl.ShadeModel(gl.SMOOTH)
	gl.LineWidth(2)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)
	gl.Hint(gl.PERSPECTIVE_CORRECTION_HINT, gl.NICEST)
	gl.ColorMaterial(gl.FRONT_AND_BACK, gl.AMBIENT_AND_DIFFUSE)
	return nil
}

// onKey handles key events.
func onKey(key, state int) {
	if key == glfw.KeyEsc {
		glfw.CloseWindow()
	}
}

// onResize handles window resize events.
func onResize(w, h int) {
	if w < 1 {
		w = 1
	}

	if h < 1 {
		h = 1
	}

	gl.Viewport(0, 0, w, h)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	glu.Perspective(45.0, float64(w)/float64(h), 0.1, 200.0)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

// Create a vertex/fragment shader program
func createSampleProgram() gl.Program {
	vs := `
#version 120
// Input vertex data, different for all executions of this shader.
// attribute vec3 vertexPosition_modelspace;
void main(){

	gl_Position = gl_Vertex;
}
	`
	fs := `
#version 120

void main()
{

	// Output color = red 
	gl_FragColor = vec4(1,0,0,1);

}
	`
	vshader := gl.CreateShader(gl.VERTEX_SHADER)
	vshader.Source(vs)
	vshader.Compile()
	if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		panic("Unable to compile vertex shader. " + vshader.GetInfoLog())
	}

	fshader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fshader.Source(fs)
	fshader.Compile()
	if fshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		panic("Unable to compile fragment shader. " + fshader.GetInfoLog())
	}

	program := gl.CreateProgram()
	program.AttachShader(vshader)
	program.AttachShader(fshader)
	program.Link()

	if program.Get(gl.LINK_STATUS) != gl.TRUE {
		panic("Unable to link program. " + fshader.GetInfoLog())
	}

	//program.Use()
	return program

}


func loadMeshInfo() {
	path, err := gas.Abs("github.com/andrebq/assimp/data/cube.dae")
	if err != nil { panic(err) }
	
	scene, err = conv.LoadAsset(path)
}