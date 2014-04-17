package main

import (
	"github.com/andrebq/exp/opala"
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
