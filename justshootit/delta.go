package main

type TimeDelta struct {
	// The initial time from the last frame
	OldTime float64
	// The time elapsed from the last call to Tick
	Delta float64
	// The function to get the initial time
	GetTime func() float64
	// The upper value that can happen between calls to Tick
	MaxDelta float64
}

func (t *TimeDelta) Tick() float64 {
	t.Delta = t.GetTime() - t.OldTime
	t.OldTime = t.GetTime()
	if t.Delta > t.MaxDelta && t.MaxDelta > 0 {
		t.Delta = t.MaxDelta
	}
	return t.Delta
}
