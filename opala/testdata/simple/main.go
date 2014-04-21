package main

import (
	"github.com/andrebq/exp/opala"
	"github.com/andrebq/gas"
	"image/color"
)

func main() {
	w, err := opala.NewDisplay(800, 600, "simple")
	if err != nil {
		panic(err)
	}

	// this create a new atlas that can later be used
	// to render images on the screen.
	backgroundLayer := opala.NewAtlas(800, 600, 1, 1)
	img, _ := backgroundLayer.AllocateDefault("img")
	pngfile := gas.MustAbs("github.com/andrebq/exp/opala/testdata/simple/f.png")
	_ = pngfile
	//img.FromPNG(gas.MustAbs(pngfile))
	img.Fill(color.RGBA{255, 0, 255, 255})

	dl := opala.NewDisplayList()
	dl.Push(w)
	opala.Vsync(true)
	for !dl.ShouldClose() {
		opala.AcquireInput()
		w.SendDraw(opala.ClearCmd{})
		w.SendDraw(&opala.DrawImage{
			Image: img,
		})
		w.Render()
	}
}
