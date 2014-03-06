package chipmunk

// #cgo LDFLAGS: -lchipmunk
// #include "chipmunk/chipmunk.h"
import "C"

type Vect C.cpVect

type Space struct {
	s *C.cpSpace
}

type Shape struct {
	shape *C.cpShape
}

type Body struct {
	body *C.cpBody
}

func f(x float32) C.cpFloat {
	return C.cpFloat(x)
}

func V(x, y float32) Vect {
	return Vect(C.cpv(f(x), f(y)))
}

func VZero() Vect {
	return V(0, 0)
}

func MomentForCircle(mass, innerDiameter, outerDiameter float32, offset Vect) float32 {
	return float32(C.cpMomentForCircle(f(mass), f(innerDiameter), f(outerDiameter), C.cpVect(offset)))
}

func NewSpace() Space {
	return Space{s: C.cpSpaceNew()}
}

func (s *Space) SetGravity(g Vect) {
	C.cpSpaceSetGravity(s.s, C.cpVect(g))
}

func (s *Space) AddShape(shape Shape) Shape {
	return Shape{
		C.cpSpaceAddShape(s.s, shape.shape)}
}

func (s *Space) AddBody(body Body) Body {
	return Body{
		C.cpSpaceAddBody(s.s, body.body)}
}

func (s *Space) StaticBody() Body {
	return Body{body: s.s.staticBody}
}

func (s *Space) Step(timeStep float32) {
	C.cpSpaceStep(s.s, f(timeStep))
}

func (s *Space) Free() {
	C.cpSpaceFree(s.s)
}

func NewSegmentShape(body Body, a, b Vect, radius float32) Shape {
	return Shape{
		shape: C.cpSegmentShapeNew(body.body, C.cpVect(a), C.cpVect(b), f(radius)),
	}
}

func NewCircleShape(body Body, radius float32, center Vect) Shape {
	return Shape{
		shape: C.cpCircleShapeNew(body.body, f(radius), C.cpVect(center)),
	}
}

func (s *Shape) SetFriction(val float32) {
	C.cpShapeSetFriction(s.shape, f(val))
}

func (s *Shape) Free() {
	C.cpShapeFree(s.shape)
}

func NewBody(mass, moment float32) Body {
	return Body{
		body: C.cpBodyNew(f(mass), f(moment))}
}

func (b *Body) SetPos(p Vect) {
	C.cpBodySetPos(b.body, C.cpVect(p))
}

func (b *Body) Pos() Vect {
	return Vect(C.cpBodyGetPos(b.body))
}

func (b *Body) Vel() Vect {
	return Vect(C.cpBodyGetVel(b.body))
}

func (b *Body) Free() {
	C.cpBodyFree(b.body)
}