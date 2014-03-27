package main

import (
	"github.com/andrebq/exp/opala"
)

func main() {
	w, err := opala.NewDisplay(800, 600, "simple")
	if err != nil {
		panic(err)
	}
	opala.Vsync(true)
	for !w.ShouldClose() {
		w.AcquireInput()
		w.Render()
	}
}
