package figo

import (
	glm "github.com/Agon/googlmath"
	"reflect"
	"testing"
)

func TestBodyBoundRect(t *testing.T) {
	rect := NewRectShape(glm.Vector2{1, 1})
	body := NewBody()

	if !reflect.DeepEqual(body.BoundRect(), NewAABB(0, 0, 0, 0)) {
		t.Errorf("Body should have a empty bound rect. got %v", body.BoundRect())
	}

	body.AddShape(rect)

	if !reflect.DeepEqual(body.BoundRect(), NewAABB(-1, -1, 1, 1)) {
		t.Errorf("Body have a invalid bound rect. got %v", body.BoundRect())
	}

	rect.Translate(glm.Vector2{1, 1})

	if !reflect.DeepEqual(body.BoundRect(), NewAABB(0, 0, 2, 2)) {
		t.Errorf("body aabb should have changed after changing the shape. got %v", body.BoundRect())
	}
}
