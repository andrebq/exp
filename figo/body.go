package figo

// Body is a collection of shapes that have their
// vertices kept in a solid formation
//
// Ie, Body is a RigidBody
type Body struct {
	shapes      []*Shape
	boundRect   AABB
	invalidRect bool
}

func NewBody() *Body {
	return &Body{
		shapes: make([]*Shape, 0, 1),
	}
}

// AddShape inserts a new shape to this body.
//
// This will force the body bounding rect to be
// recalculated
func (b *Body) AddShape(s *Shape) *Body {
	if s.body != nil {
		panic("a shape cannot be attached to another body. call body.RemoveShape first")
	}
	s.body = b
	b.shapes = append(b.shapes, s)
	b.invalidRect = true
	return b
}

// BoundRect returns the axis aligned bouding rectangle
// for this body.
//
// The returned rectanble is large enough to hold all
// shapes that form this body
func (b *Body) BoundRect() AABB {
	if b.invalidRect {
		b.recalculateBoundRect()
	}
	return b.boundRect
}

// recalculateBoundRect will scan all shapes and make
// a rectangle that is large enough to hold all shapes
func (b *Body) recalculateBoundRect() {
	b.boundRect.Set(0, 0, 0, 0)
	b.invalidRect = false
	if len(b.shapes) == 0 {
		return
	}
	for _, s := range b.shapes {
		aabb := s.AABB()
		b.boundRect.MergeWith(&aabb)
	}
}
