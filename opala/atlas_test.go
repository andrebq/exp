package opala

import (
	"fmt"
	glm "github.com/Agon/googlmath"
	"reflect"
	"testing"
)

func TestAtlasAllocate(t *testing.T) {
	// a small atlas with one single chunk
	a := NewAtlas(32, 32, 1, 1)
	if v, err := a.AllocateDefault("one"); err != nil {
		t.Errorf("Atlas is large enough, should allocate one")
		same, _ := a.AllocateDefault("one")
		if same != v {
			t.Errorf("same chunk name should give same pointer")
		}

		same = a.ChunkAt(0, 0)
		if same != v {
			t.Errorf("chunk at 0,0 should exist")
		}
	}
	if _, err := a.AllocateDefault("two"); err == nil {
		t.Errorf("Atlas is too small, shouldn't allocate two")
	}

	a = NewAtlas(32, 32, 2, 2)
	// this should give 4 consecutive allocations
	// and then stop
	for i := 0; i < 4; i++ {
		_, err := a.AllocateDefault(fmt.Sprintf("c%v", i))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
	if _, err := a.AllocateDefault("five"); err == nil {
		t.Errorf("Atlas is too small, shouldn't allocate five")
	}
}

func TestAtlasUVRect(t *testing.T) {
	// here is the example
	// we have a 2x2 grid of 32x32 images,
	// and we want to map it to opengl
	//
	// the image at 0,0 should receive a uv rect of
	// (0,0.5) (0.5, 1)
	//
	// and the image at 1,0 should receive a uv rect
	// of (0,0) (0.5, 0.5)
	uvr00 := UVRect{
		BottomLeft: glm.Vector2{
			X: 0,
			Y: 0.5,
		},
		TopRight: glm.Vector2{
			X: 0.5,
			Y: 1,
		},
	}

	uvr10 := UVRect{
		BottomLeft: glm.Vector2{
			X: 0,
			Y: 0,
		},
		TopRight: glm.Vector2{
			X: 0.5,
			Y: 0.5,
		},
	}

	a := NewAtlas(32, 32, 2, 2)
	a.AllocateMany("uv", 4)

	c00 := a.ChunkAt(0, 0)
	c10 := a.ChunkAt(1, 0)

	if c00 == nil || c10 == nil {
		t.Fatalf("chunks shouldn't be nil")
	}

	if !reflect.DeepEqual(uvr00, c00.UVRect()) {
		t.Errorf("c00 should have uv %v but got %v", uvr00, c00.UVRect())
	}

	if !reflect.DeepEqual(uvr10, c10.UVRect()) {
		t.Errorf("c00 should have uv %v but got %v", uvr00, c00.UVRect())
	}
}

func TestNonPowerOf2(t *testing.T) {
	cw, aw := float32(800), float32(1024)
	ch, ah := float32(600), float32(1024)

	uvr00 := UVRect{
		BottomLeft: glm.Vector2{
			X: 0,
			Y: 0,
		},
		TopRight: glm.Vector2{
			X: cw / aw,
			Y: ch / ah,
		},
	}

	at := NewAtlas(int(cw), int(ch), 1, 1)
	dx, dy := float32(at.data.Bounds().Dx()), float32(at.data.Bounds().Dy())
	if dx != aw || dy != ah {
		t.Fatalf("invalid size. expected %v,%v got ¨%v,%v", aw, ah, dx, dy)
	}
	_ = uvr00

	chunk, _ := at.AllocateDefault("uv")

	rect := chunk.UVRect()

	if !reflect.DeepEqual(rect, uvr00) {
		t.Fatalf("expecting: %v got %v", uvr00, rect)
	}
}
