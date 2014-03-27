package opala

import (
	"fmt"
	glm "github.com/Agon/googlmath"
	"image"
)

// UVRect represent the coordinate system for
// used by OpenGL
type UVRect struct {
	BottomLeft glm.Vector2
	TopRight   glm.Vector2
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
}

// This will create a new atlas with enough space to
// nRows X nColumns chunks of widthXheight size
func NewAtlas(width, height, nRows, nColumns int) *Atlas {
	return &Atlas{
		cw:   width,
		ch:   height,
		data: image.NewRGBA(image.Rect(0, 0, nRows*height, nColumns*width)),
	}
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

func (a *Atlas) calculateUvFor(c *AtlasChunk) UVRect {
	rows, cols := a.gridSize()
	imgWidth, imgHeight := float32(cols*a.cw), float32(rows*a.ch)
	// since in opengl the 0,0 means the bottomleft and
	// image.RGBA 0,0 means topleft
	// we need to offset cY by the heigth of each chunk
	cX, cY := float32(c.column*a.cw), float32(c.row*a.ch+a.ch)
	uvr := UVRect{
		BottomLeft: glm.Vector2{
			X: cX / imgWidth,
			Y: 1 - cY/imgHeight,
		},
	}
	uvr.TopRight.X = uvr.BottomLeft.X + float32(a.cw)/imgWidth
	uvr.TopRight.Y = uvr.BottomLeft.Y + float32(a.ch)/imgHeight
	return uvr
}

func (a *Atlas) subImage(c *AtlasChunk) *image.RGBA {
	return nil
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
		}
	}
	last := a.chunks[len(a.chunks)-1]
	next := &AtlasChunk{
		row:    last.row,
		column: last.column + 1,
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
	subdata     *image.RGBA
	name        string
	row, column int
	uvrect      UVRect
}

func (ic *AtlasChunk) Size() (w, h int) {
	rect := ic.subdata.Bounds()
	return rect.Dx(), rect.Dy()
}

func (ic *AtlasChunk) UVRect() UVRect {
	return ic.uvrect
}

func (ic *AtlasChunk) String() string {
	return fmt.Sprintf("%v [%v,%v] uv[%v,%v]",
		ic.name,
		ic.row,
		ic.column,
		ic.uvrect.BottomLeft,
		ic.uvrect.TopRight)
}
