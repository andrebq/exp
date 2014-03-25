package opala

import (
	"fmt"
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
