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

// MinMaxVec will return a pair of vectors hold the smallest (x,y)
// pair from a and b. And the biggest (x,y) pair from a and b.
//
// TODO: put a example here
//
// min, max := MinMaxVec(glm.Vector2{-1, 10}, glm.Vector2{10, -1})
//
// min == glm.Vector2{-1, -1}
// max == glm.Vector2{10, 10}
func MinMaxVec(a, b glm.Vector2) (min glm.Vector2, max glm.Vector2) {
	min.X = glm.Min(a.X, b.X)
	min.Y = glm.Min(a.Y, b.Y)

	max.X = glm.Max(a.X, b.X)
	max.Y = glm.Max(a.Y, b.Y)
	return
}
