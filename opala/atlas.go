package opala

import (
	"fmt"
	glm "github.com/Agon/googlmath"
	"github.com/go-gl/gl"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
)

// UVRect represent the coordinate system for
// used by OpenGL
type UVRect struct {
	// TopLeft (TL), TopRigth (TR), BottomLeft (BL), BottomRight (BR)
	TL, TR, BL, BR glm.Vector2
}

// Atlas holds a large texture in memory and enable users
// to ask for chunks of the image.
//
// The size of each chunk is fixed, this might require more
// atlases to be created bug makes the code here much more easier
// to understand.
//
// TODO: maybe this should be more generic and work with other
// image formats, but at this moment just image.RGBA is supported
type Atlas struct {
	// main data
	data *image.RGBA
	// allocated chunks
	chunks []*AtlasChunk
	// Chunk width and height
	cw, ch int

	gltex gl.Texture
}

// NewAtlas will create a new atlas with enough space to
// at least nRows X nColumns chunks of widthXheight size. The total
// number of chunks are determinated by the power of two rule below.
//
// The actual memory used might be larger since all atlas MUST BE
// a power of 2 rect (16, 32, 64, 128...). But the rectangle used
// by the chunk is limited to width/height.
func NewAtlas(width, height, nRows, nColumns int) *Atlas {
	_, max := minMaxOf(powerOfTwo(nRows*height), powerOfTwo(nColumns*width))
	return &Atlas{
		cw:    width,
		ch:    height,
		data:  image.NewRGBA(image.Rect(0, 0, max, max)),
		gltex: gl.Texture(gl.FALSE),
	}
}

func powerOfTwo(val int) int {
	val = int(math.Pow(2, math.Ceil(math.Log2(float64(val)))))
	return val
}

func minMaxOf(values ...int) (int, int) {
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// AllocateMany will try to allocate at most count chunks
// and return the slice with allocated chunks.
//
// If an error happens the slice will hold the chunks that
// could be allocated before the error.
//
// Chunks are named by concatenating prefix + "-" + index
func (a *Atlas) AllocateMany(prefix string, count int) ([]*AtlasChunk, error) {
	if count < 0 {
		return nil, CountMustBePositive
	}
	ret := make([]*AtlasChunk, 0, count)
	for i := 0; i < count; i++ {
		c, err := a.AllocateDefault(fmt.Sprintf("%v-%v", prefix, i))
		if err != nil {
			return ret, err
		}
		ret = append(ret, c)
	}
	return ret, nil
}

func (a *Atlas) AllocateDefault(name string) (*AtlasChunk, error) {
	if v := a.findByName(name); v != nil {
		return v, nil
	}
	ci := a.emptyChunk()
	if ci == nil {
		return nil, AtlasIsFull
	}
	ci.name = name
	ci.subdata = a.subImage(ci)
	ci.uvrect = a.calculateUvFor(ci)
	a.chunks = append(a.chunks, ci)
	return ci, nil
}

// ChunkAt returns the chunk at the given row and column
// if no chunk is found, returns nil
func (a *Atlas) ChunkAt(row, column int) *AtlasChunk {
	for _, v := range a.chunks {
		if v.row == row && v.column == column {
			return v
		}
	}
	return nil
}

func (a *Atlas) unbind(release bool) error {
	if !gl.Object(a.gltex).IsTexture() {
		return nil
	}
	a.gltex.Unbind(gl.TEXTURE_2D)
	if release {
		a.gltex.Delete()
	}
	a.gltex = gl.Texture(gl.FALSE)
	return checkGlError()
}

// bind the given atlas to the current GL context
//
// the current implementation is very stupid, since it will
// upload the texture every single call.
//
// later, improve this to upload only if there is a real need for it
func (a *Atlas) bind() error {
	// discard any possible error
	if err := checkGlError(); err != nil {
		return err
	}
	if gl.Object(a.gltex).IsTexture() {
		a.gltex = gl.GenTexture()
	}
	a.gltex.Bind(gl.TEXTURE_2D)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, a.data.Bounds().Dx(), a.data.Bounds().Dy(), 0, gl.RGBA, gl.UNSIGNED_BYTE, a.data.Pix)
	if err := checkGlError(gl.OUT_OF_MEMORY, gl.INVALID_OPERATION); err != nil {
		return err
	}

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.GenerateMipmap(gl.TEXTURE_2D)
	panicGlError()
	return nil
}

func (a *Atlas) calculateUvFor(c *AtlasChunk) UVRect {
	uvr := UVRect{}
	bounds := a.data.Bounds()
	x0, y0 := c.column*a.ch, c.row*a.cw
	x1, y1 := x0, y0+a.ch
	x2, y2 := x0+a.cw, y0
	// here x1, y1 is the bottom-left
	// and x2, y2 is the top-rigth
	// considering that 0,0 is top-left
	// in opengl case, top-left is 0, Height
	// to convert, just remove the Height from y1 and y2
	y0, y1, y2 = bounds.Dy()-y0, bounds.Dy()-y1, bounds.Dy()-y2

	// since uv must range from 0, 1
	// and the image range from 0 to width/height
	// we need a scale, ie, normalize the vector
	uvr.TL.X, uvr.TL.Y = float32(x1), float32(y2)
	uvr.TR.X, uvr.TR.Y = float32(x2), float32(y2)
	uvr.BL.X, uvr.BL.Y = float32(x1), float32(y1)
	uvr.BR.X, uvr.BR.Y = float32(x2), float32(y1)
	X, Y := float32(bounds.Dx()), float32(bounds.Dy())
	divideVectorBy(&uvr.TL, X, Y)
	divideVectorBy(&uvr.TR, X, Y)
	divideVectorBy(&uvr.BL, X, Y)
	divideVectorBy(&uvr.BR, X, Y)
	return uvr
}

func divideVectorBy(v *glm.Vector2, x, y float32) {
	v.X = v.X / x
	v.Y = v.Y / y
}

func (a *Atlas) subImage(c *AtlasChunk) *image.RGBA {
	x0, y0 := a.cw*c.column, a.ch*c.row
	x1, y1 := x0+a.cw, y0+a.cw
	return a.data.SubImage(image.Rect(x0, y0, x1, y1)).(*image.RGBA)
}

func (a *Atlas) findByName(name string) *AtlasChunk {
	for _, v := range a.chunks {
		if v.name == name {
			return v
		}
	}
	return nil
}

func (a *Atlas) emptyChunk() *AtlasChunk {
	if len(a.chunks) == 0 {
		return &AtlasChunk{
			row:    0,
			column: 0,
			atlas:  a,
		}
	}
	last := a.chunks[len(a.chunks)-1]
	next := &AtlasChunk{
		row:    last.row,
		column: last.column + 1,
		atlas:  a,
	}
	// check if we can use the same row
	// and the next column
	if a.validChunk(next) {
		return next
	}
	next.row++
	next.column = 0
	// check if we can use the next row
	// and the first column
	if a.validChunk(next) {
		return next
	}
	// full
	return nil
}

func (a *Atlas) validChunk(chunk *AtlasChunk) bool {
	rows, columns := a.gridSize()
	return chunk.row < rows && chunk.column < columns
}

// Size of the chunck grid, this is
// the size of the image divided by the size
// of each chunk
func (a *Atlas) gridSize() (rows, cols int) {
	rect := a.data.Bounds()
	width, height := rect.Dx(), rect.Dy()
	return height / a.ch, width / a.cw
}

// hold the information about a chunk of an atlas
type AtlasChunk struct {
	atlas       *Atlas
	subdata     *image.RGBA
	name        string
	row, column int
	uvrect      UVRect
}

func (ic *AtlasChunk) Size() (w, h int) {
	rect := ic.subdata.Bounds()
	return rect.Dx(), rect.Dy()
}

func (ic *AtlasChunk) Sizef() (w, h float32) {
	iw, ih := ic.Size()
	return float32(iw), float32(ih)
}

func (ic *AtlasChunk) UVRect() UVRect {
	return ic.uvrect
}

func (ic *AtlasChunk) Fill(c color.Color) {
	u := image.NewUniform(c)
	draw.Draw(ic.subdata, ic.subdata.Bounds(), u, image.Point{0, 0}, draw.Over)
}

func (ic *AtlasChunk) FromPNG(file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()
	img, err := png.Decode(fd)
	if err != nil {
		return err
	}
	draw.Draw(ic.subdata, ic.subdata.Bounds(), img, image.Point{0, 0}, draw.Over)
	return nil
}

func (ic *AtlasChunk) String() string {
	return fmt.Sprintf("%v [%v,%v] uv[%v,%v]",
		ic.name,
		ic.row,
		ic.column,
		ic.uvrect)
}
