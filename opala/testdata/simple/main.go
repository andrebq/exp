package main

import (
	"github.com/andrebq/exp/opala"
)

func main() {
	w, err := opala.NewDisplay(800, 600, "simple")
	if err != nil {
		panic(err)
	}

	other, err := opala.NewDisplay(800, 600, "other")
	if err != nil {
		panic(err)
	}
	dl := opala.NewDisplayList()
	dl.Push(w, other)
	opala.Vsync(true)
	for !dl.ShouldClose() {
		opala.AcquireInput()
		dl.Render()
	}
}
