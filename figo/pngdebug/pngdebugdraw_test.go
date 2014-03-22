package pngdebug

import (
	glm "github.com/Agon/googlmath"
	"github.com/andrebq/exp/figo"
	"testing"
)

func TestDrawShape(t *testing.T) {
	dbg := &PngDebugDraw{
		TimeBetweenFrames: 0,
		OutputFolder:      "pngdebugdraw-out",
		DecToPix:          1,
		SaveToDisk:        true,
	}

	dbg.BeginFrame(figo.FrameData{HalfWidth: 500, HalfHeight: 500})
	shape1 := figo.NewRectShape(glm.Vector2{10, 10})
	// draw a static shape
	dbg.DrawShape(shape1, 10, 0, 0)

	shape2 := shape1.Clone()
	// draw a moving shape
	shape2.Translate(glm.Vector2{30, 0})
	dbg.DrawShape(shape2, 10, 10, 0)

	shape3 := shape2.Clone()
	// draw a moving and accelerating shape
	shape3.Translate(glm.Vector2{30, 0})
	dbg.DrawShape(shape3, 10, 10, 10)

	dbg.EndFrame()
}
