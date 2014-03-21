package figo

import (
	glm "github.com/Agon/googlmath"
)

// VecToFloat convert a input slice of Vector2 objects
// to a flat slice of float32 values.
//
// If out is large enough (cap(out) >= len(in)*2) then no
// allocation is done.
//
// If out is smaller or nil, then a new slice is allocated
func VecToFloat(out []float32, in ...glm.Vector2) []float32 {
	if cap(out) < len(in)*2 {
		out = make([]float32, len(in)*2)
	} else {
		out = out[:len(in)*2]
	}

	for i, v := range in {
		out[i], out[i+1] = v.X, v.Y
	}
	return out
}

func FloatToVec(out []glm.Vector2, in ...float32) []glm.Vector2 {
	if len(in)%2 != 0 {
		panic("input length must be divisible by 2")
	}
	if cap(out)*2 < len(in) {
		out = make([]glm.Vector2, len(in)/2)
	} else {
		out = out[:len(in)/2]
	}
	outPos := 0
	inPos := 0
	for outPos*2 < len(in) {
		out[outPos].X, out[outPos].Y = in[inPos], in[inPos+1]
		outPos++
		inPos += 2
	}
	return out
}
