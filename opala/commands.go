package opala

import (
	"fmt"
	glm "github.com/Agon/googlmath"
	"github.com/andrebq/gas"
	"github.com/go-gl/gl"
	"github.com/go-gl/glh"
	"io/ioutil"
	"path/filepath"
)

type ClearCmd struct{}

func (c ClearCmd) Render(d *Display) error {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	return checkGlError()
}

func (c ClearCmd) Name() string {
	return "clear-cmd"
}

type DrawCmd interface {
	Name() string
	Render(d *Display) error
}

type DrawImage struct {
	Image   *AtlasChunk
	buf     *glh.MeshBuffer
	program gl.Program
}

func (d *DrawImage) Name() string { return "draw-image" }

func (d *DrawImage) Render(display *Display) error {
	if err := display.bindAtlas(d.Image.atlas); err != nil {
		return err
	}
	d.rebuildBuffer()
	if err := d.rebuildProgram(); err != nil {
		return err
	}
	var err error
	defer func() {
		if err == nil {
			err = d.clear()
		} else {
			d.clear()
		}
	}()
	if err = d.setUniforms(); err != nil {
		return err
	}
	if err = d.render(); err != nil {
		return err
	}
	return err
}

func (d *DrawImage) setUniforms() error {
	p := d.program
	gl.ActiveTexture(gl.TEXTURE0)
	d.Image.atlas.bind()
	if err := checkGlError(); err != nil {
		return err
	}

	loc := p.GetUniformLocation("mysample")
	loc.Uniform1i(0)

	return checkGlError()
}

func (d *DrawImage) render() error {
	d.buf.Render(gl.TRIANGLES)
	return checkGlError()
}

func (d *DrawImage) clear() error {
	d.buf.Clear()
	d.program.Delete()
	return checkGlError()
}

func (d *DrawImage) rebuildProgram() error {
	file, err := gas.Abs("github.com/andrebq/exp/opala/shaders")
	if err != nil {
		return err
	}
	program, err := compileProgram(file, "renderimage")
	if err != nil {
		return err
	}
	d.program = program
	return checkGlError()
}

func (d *DrawImage) rebuildBuffer() {
	buf := glh.NewMeshBuffer(
		glh.RenderArrays,

		glh.NewPositionAttr(3, gl.FLOAT, gl.STATIC_DRAW))

	vertices := scale(0.5, []float32{
		-1, -1, 0,
		-1, 1, 0,
		1, 1, 0,
	})
	buf.Add(vertices)

	vertices = scale(0.5, []float32{
		1, 1, 0,
		1, -1, 0,
		-1, -1, 0,
	})
	buf.Add(vertices)
	d.buf = buf
}

func scale(scale float32, in []float32) []float32 {
	for i, v := range in {
		in[i] = v * scale
	}
	return in
}

func norm(a, b float32) (float32, float32) {
	vec := glm.Vector2{a, b}
	vec = vec.Nor()
	return vec.X, vec.Y
}

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
		return glProg, fmt.Errorf("Unable to link glProg. Cause: %v", glProg.GetInfoLog())
	}
	glProg.Use()

	return
}
