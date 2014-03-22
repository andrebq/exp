package figo

import (
	glm "github.com/Agon/googlmath"
	"reflect"
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

func TestCloneShape(t *testing.T) {
	shape := NewRectShape(glm.Vector2{1, 1})
	shape.Translate(glm.Vector2{10, 10})

	clone := shape.Clone()

	if !reflect.DeepEqual(shape.Vertices(), clone.Vertices()) {
		t.Errorf("unequal vertices. expected %v got %v", shape.Vertices(), clone.Vertices())
	}

	if !reflect.DeepEqual(shape.TransformedVertices(), clone.TransformedVertices()) {
		t.Errorf("unequal world vertices. expected %v got %v", shape.TransformedVertices(), clone.TransformedVertices())
	}
}
