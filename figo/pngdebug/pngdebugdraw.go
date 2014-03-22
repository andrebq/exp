package pngdebug

import (
	"code.google.com/p/draw2d/draw2d"
	"fmt"
	glm "github.com/Agon/googlmath"
	"github.com/andrebq/exp/figo"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

// Implement the debugdraw interface from figo
// and save the output to a gif file
type PngDebugDraw struct {
	// Folder to store the frames
	OutputFolder string
	// Decimal to Pixel conversion factor
	//
	// The engine works on decimal, so 1 means 1 meter
	// this is used to convert 1 meter to N pixels
	DecToPix float32
	// If the next frame should be saved to disk when
	// EndFrame is called
	//
	// Usefull to skip frames where the visual change is
	// small
	SaveToDisk bool
	// Total time from the simulation (fixed time step) that
	// should be waited to save a new frame.
	//
	// If dt = 1/60, and you want to save at a 1/30 fps you should
	// set this to 1/30 then the simulation will be saved every 2 frames
	//
	// A value of 0 means save every frame
	TimeBetweenFrames float32

	timeToNextFrame float32
	cImg            draw.Image
	ctx             *draw2d.ImageGraphicContext
	offset          glm.Vector2
	frame           figo.FrameData
}

func (f *PngDebugDraw) BeginFrame(info figo.FrameData) {
	f.frame = info
	f.cImg = image.NewRGBA(WorldRect(info.HalfWidth, info.HalfHeight, f.DecToPix))
	f.ctx = draw2d.NewGraphicContext(f.cImg)
	f.offset = glm.Vector2{info.HalfWidth, info.HalfHeight}
	f.ctx.SetFillColor(color.White)
	f.ctx.Clear()
}

func (s *PngDebugDraw) DrawShape(shape *figo.Shape, mass, vel, acc float32) {
	ps := draw2d.NewPathStorage()
	data := shape.TransformedVertices()
	_x, _y := 0.0, 0.0
	for i := 0; i < len(data); i += 2 {
		x := float64((data[i] + s.offset.X) * s.DecToPix)
		y := float64((data[i+1] + s.offset.Y) * s.DecToPix)
		if i == 0 {
			_x, _y = x, y
			ps.MoveTo(x, y)
		} else {
			ps.LineTo(x, y)
		}
	}
	ps.LineTo(_x, _y)
	s.ctx.SetFillColor(color.RGBA{R: 255, A: 255})
	s.ctx.FillStroke(ps)
}

func (s *PngDebugDraw) EndFrame() {
	if !s.SaveToDisk {
		return
	}
	if s.timeToNextFrame > 0 {
		s.timeToNextFrame -= s.frame.FixedDelta
		return
	}
	s.timeToNextFrame = s.TimeBetweenFrames
	fileName := filepath.Join(s.OutputFolder, fmt.Sprintf("./png-frame-%v.png", s.frame.Number))
	os.MkdirAll(s.OutputFolder, 0600)
	os.Remove(fileName)
	file, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer file.Close()
	png.Encode(file, s.cImg)
}

func WorldRect(hw, hh, decToPix float32) image.Rectangle {
	ret := image.Rect(
		0,
		0,
		int(hw*decToPix*2),
		int(hh*decToPix*2))
	return ret
}
