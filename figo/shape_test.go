package figo

import (
	glm "github.com/Agon/googlmath"
	"testing"
)

func TestShapeSetAs(t *testing.T) {
	shape := &Shape{}
	shape.SetAsRect(glm.Vector2{1, 1})

	expected := []glm.Vector2{
		glm.Vector2{-1, 1},
		glm.Vector2{1, 1},
		glm.Vector2{1, -1},
		glm.Vector2{-1, -1},
	}

	if len(expected) != len(shape.Vertex) {
		t.Errorf("Length should be %v but it is %v", len(expected),
			len(shape.Vertex))
	} else {
		for i, v := range shape.Vertex {
			if expected[i] != v {
				t.Errorf("Vertex at [%v] should be %v but got %v",
					i, expected[i], v)
			}
		}
	}
}
