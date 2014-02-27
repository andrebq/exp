package main

import (
	glm "github.com/Agon/googlmath"
)

type Mesh []Vec3

func (m *Mesh) Push(v Vec3) {
	*m = append(*m, v)
}

func (m *Mesh) Add(v Vec3) {
	for i, _ := range *m {
		m[i] = m[i].Add(v)
	}
}
