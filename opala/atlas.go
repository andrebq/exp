package opala

import (
	"image"
)

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
	a.chunks = append(a.chunks, ci)
	return ci, nil
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
}

func (ic *AtlasChunk) Size() (w, h int) {
	rect := ic.subdata.Bounds()
	return rect.Dx(), rect.Dy()
}
