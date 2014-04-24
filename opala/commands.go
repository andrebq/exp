package opala

import (
	"fmt"
	glm "github.com/Agon/googlmath"
	"github.com/andrebq/gas"
	"github.com/go-gl/gl"
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
	mvp     *glm.Matrix4
	program gl.Program
	vertex  [][]float32
	uv      [][]float32
}

func (d *DrawImage) Name() string { return "draw-image" }

func (d *DrawImage) Render(display *Display) error {
	if err := display.bindAtlas(d.Image.atlas); err != nil {
		return err
	}
	d.rebuildModel(display)
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
	panicGlError()

	loc := p.GetUniformLocation("mysample")
	loc.Uniform1i(0)
	panicGlError()

	worldMatrix := p.GetUniformLocation("MVP")
	worldMatrix.UniformMatrix4f(false, ptrForMatrix(d.mvp))
	panicGlError()
	return nil
}

func (d *DrawImage) render() error {
	fn := func(idx int) (err error) {
		defer func() {
			//err = recover()
		}()
		vbuf := gl.GenBuffer()
		vbuf.Bind(gl.ARRAY_BUFFER)
		gl.BufferData(gl.ARRAY_BUFFER, len(d.vertex[idx])*4, d.vertex[idx], gl.STATIC_DRAW)
		panicGlError()
		defer vbuf.Delete()

		uvbuf := gl.GenBuffer()
		uvbuf.Bind(gl.ARRAY_BUFFER)
		gl.BufferData(gl.ARRAY_BUFFER, len(d.uv[idx])*4, d.uv[idx], gl.STATIC_DRAW)
		panicGlError()
		defer uvbuf.Delete()

		vloc := gl.AttribLocation(0)
		vloc.EnableArray()
		vbuf.Bind(gl.ARRAY_BUFFER)
		vloc.AttribPointer(3, gl.FLOAT, false, 0, nil)
		panicGlError()
		defer vloc.DisableArray()

		uvloc := gl.AttribLocation(1)
		uvloc.EnableArray()
		uvbuf.Bind(gl.ARRAY_BUFFER)
		uvloc.AttribPointer(2, gl.FLOAT, false, 0, nil)
		panicGlError()
		defer uvloc.DisableArray()

		gl.DrawArrays(gl.TRIANGLES, 0, 3)

		return checkGlError()
	}
	if err := fn(0); err != nil {
		return err
	}
	return fn(1)
}

func (d *DrawImage) clear() error {
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

func (d *DrawImage) rebuildModel(display *Display) {
	if len(d.vertex) != 2 {
		d.vertex = make([][]float32, 2)
		d.uv = make([][]float32, 2)
	}

	modelM := glm.NewIdentityMatrix4()
	d.mvp = display.vp.Mul(modelM)

	d.vertex[0] = scale(10, []float32{
		-1, -1, 0,
		-1, 1, 0,
		1, 1, 0,
	})
	d.uv[0] = []float32{
		1, 0.5,
		1, 0.5,
		1, 0.5,
	}

	d.vertex[1] = scale(10, []float32{
		1, 1, 0,
		1, -1, 0,
		-1, -1, 0,
	})
	d.uv[1] = []float32{
		1, 0.5,
		1, 0.5,
		1, 0.5,
	}
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

func ptrForMatrix(m *glm.Matrix4) *[16]float32 {
	ret := [16]float32{
		m.M11, m.M12, m.M13, m.M14,
		m.M21, m.M22, m.M23, m.M24,
		m.M31, m.M32, m.M33, m.M34,
		m.M41, m.M42, m.M43, m.M44,
	}
	return &ret
}
