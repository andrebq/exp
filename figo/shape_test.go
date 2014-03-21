package figo

import (
	glm "github.com/Agon/googlmath"
	"testing"
)

func TestShapeSetAs(t *testing.T) {
	shape := NewRectShape(glm.Vector2{1, 1})

	expected := []glm.Vector2{
		glm.Vector2{-1, 1},
		glm.Vector2{1, 1},
		glm.Vector2{1, -1},
		glm.Vector2{-1, -1},
	}

	actual := FloatToVec(nil, shape.Vertices()...)

	if len(expected) != len(actual) {
		t.Errorf("Length should be %v but it is %v", len(expected),
			len(actual))
	} else {
		for i, v := range actual {
			if expected[i] != v {
				t.Errorf("Vertex at [%v] should be %v but got %v",
					i, expected[i], v)
			}
		}
	}
}
