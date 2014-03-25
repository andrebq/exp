package figo

import (
	glm "github.com/Agon/googlmath"
)

// A axis-aligned bounding box
type AABB struct {
	Min glm.Vector2
	Max glm.Vector2
}

func NewAABB(minX, minY, maxX, maxY float32) AABB {
	return AABB{
		Min: glm.Vector2{minX, minY},
		Max: glm.Vector2{maxX, maxY},
	}
}

// MergeWith will expand this AABB to contain the other AABB
func (a *AABB) MergeWith(other *AABB) {
	min, _ := MinMaxVec(a.Min, other.Min)
	_, max := MinMaxVec(a.Max, other.Max)
	a.Min.X = min.X
	a.Min.Y = min.Y

	a.Max.X = max.X
	a.Max.Y = max.Y
}

func (a *AABB) Set(minX, minY, maxX, maxY float32) {
	a.Min.X, a.Min.Y = minX, minY
	a.Max.X, a.Max.Y = maxX, maxY
}
