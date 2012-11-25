package main
/*Copyright (c) 2012 André Luiz Alves Moraes

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.*/

import (
	"errors"
	"flag"
	"github.com/andrebq/assimp"
	"github.com/andrebq/assimp/conv"
	"github.com/andrebq/gas"
	"github.com/go-gl/gl"
	"github.com/go-gl/glfw"
	"github.com/go-gl/glh"
	"github.com/go-gl/glu"
	"io/ioutil"
	"log"
	"os"
	"time"
	"unicode/utf8"
)

type shaderInfo struct {
	shaderName string
	fragMod    time.Time
	vertMod    time.Time
	vertCode   string
	fragCode   string
}

func loadShaderInfo(name string) (*shaderInfo, error) {
	si := &shaderInfo{}
	vname := name + ".vertex.glsl"
	fname := name + ".frag.glsl"

	vstat, err := os.Stat(vname)
	if err != nil {
		return nil, err
	}
	si.vertMod = vstat.ModTime()

	fstat, err := os.Stat(fname)
	if err != nil {
		return nil, err
	}
	si.fragMod = fstat.ModTime()

	vcode, err := ioutil.ReadFile(vname)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(vcode) {
		return nil, errors.New("Vertex shader must be utf-8")
	}

	fcode, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(fcode) {
		return nil, errors.New("Fragment shader must be utf-8")
	}
	si.shaderName = name
	si.vertCode = string(vcode)
	si.fragCode = string(fcode)

	return si, nil
}

func loadShaderInfoIfNew(name string, vtime, ftime time.Time) (*shaderInfo, error) {
	vname := name + ".vertex.glsl"
	fname := name + ".frag.glsl"

	vstat, err := os.Stat(vname)
	if err != nil {
		return nil, err
	}

	fstat, err := os.Stat(fname)
	if err != nil {
		return nil, err
	}

	if vstat.ModTime().After(vtime) || fstat.ModTime().After(ftime) {
		return loadShaderInfo(name)
	}
	return nil, nil
}

var (
	scene    *assimp.Scene
	meshFile = flag.String("if", "", "Sample cube")
	lastErr  error
)

func main() {
	flag.Parse()

	loadMeshInfo()

	err := initGL()
	if err != nil {
		log.Printf("InitGL: %v", err)
		return
	}

	program, shaderInfo, err := loadShaders(gl.Program(0), nil)
	if err != nil {
		panic(err)
	}
	_ = program

	defer glfw.Terminate()

	mb := createBuffer()
	defer mb.Release()

	// Perform the rendering.
	var angle float32
	reload := time.Tick(time.Duration(300 * time.Millisecond))
	for glfw.WindowParam(glfw.Opened) > 0 {
		select {
		case <-reload:
			oldInfo := shaderInfo
			program, shaderInfo, err = loadShaders(program, shaderInfo)
			if err != nil && lastErr == nil {
				lastErr = err
				println("Error loading shaders. Using old code.", lastErr)
			} else if err == nil && oldInfo != shaderInfo {
				lastErr = nil
				println("new shader code loaded")
			}
		default:
			// do nothing here
		}
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.LoadIdentity()
		gl.Translatef(0, 0, -20)
		gl.Rotatef(angle, 1, 1, 1)
		program.Use()

		gl.Enable(gl.COLOR_MATERIAL)
		gl.Enable(gl.POLYGON_OFFSET_FILL)
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		mb.Render(gl.TRIANGLES)

		angle += 0.5
		glfw.SwapBuffers()
	}
}

// Create the glh.MeshBufer
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

	gl.ClearColor(0.2, 0.2, 0.2, 1.0)
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
func createProgram(info *shaderInfo) (gl.Program, error) {
	vshader := gl.CreateShader(gl.VERTEX_SHADER)
	vshader.Source(info.vertCode)
	vshader.Compile()
	if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		return gl.Program(0), errors.New("Unable to compile vertex shader. " + vshader.GetInfoLog())
	}
	defer vshader.Delete() // no need to use it after linking

	fshader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fshader.Source(info.fragCode)
	fshader.Compile()
	if fshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		return gl.Program(0), errors.New("Unable to compile fragment shader. " + fshader.GetInfoLog())
	}
	defer fshader.Delete() // no need to use it after linking

	program := gl.CreateProgram()
	program.AttachShader(vshader)
	program.AttachShader(fshader)
	program.Link()

	if program.Get(gl.LINK_STATUS) != gl.TRUE {
		return gl.Program(0), errors.New("Unable to link program. " + fshader.GetInfoLog())
	}

	return program, nil
}

// Load the mesh information.
func loadMeshInfo() {

	println(*meshFile)
	if *meshFile == "" {
		path, err := gas.Abs("github.com/andrebq/assimp/data/cube.dae")
		if err != nil {
			panic(err)
		}
		*meshFile = path
	}
	println(*meshFile)

	var err error
	scene, err = conv.LoadAsset(*meshFile)
	if err != nil {
		panic(err)
	}
}

// Load the shader only if it have been modified
func loadShaders(oldProgram gl.Program, last *shaderInfo) (gl.Program, *shaderInfo, error) {
	if last != nil {
		newInfo, err := loadShaderInfoIfNew("sample", last.vertMod, last.fragMod)
		if err != nil {
			return oldProgram, last, err
		}
		if newInfo == nil {
			// nothing new, can reuse the old code
			return oldProgram, last, nil
		} else {
			newProgram, err := createProgram(newInfo)
			if err != nil {
				return oldProgram, last, err
			}
			oldProgram.Delete()
			return newProgram, newInfo, err
		}
	}
	last, err := loadShaderInfo("sample")
	if err != nil {
		return oldProgram, last, err
	}
	oldProgram, err = createProgram(last)

	return oldProgram, last, err
}
