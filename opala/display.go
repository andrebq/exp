package opala

import (
	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"log"
	"runtime"
	"sync"
)

func logGlfwError(err glfw.ErrorCode, desc string) {
	log.Printf("[glfw] %v: %v", err, desc)
}

func init() {
	runtime.LockOSThread()
}

// Tick holds information about display rendering times
type Tick struct {
	lastTime       float64
	lastRenderTime float64
	delta          float64
	realDelta      float64
	delta32        float32
	maxDelta       float64
}

// Move the timers using glfw.GetTime function.
//
// The delta is capped at maxDelta, but realDelta
// will hold the uncapped delta
func (t *Tick) tick() {
	now := glfw.GetTime()
	t.realDelta = now - t.lastTime
	t.delta = t.realDelta
	if t.delta > t.maxDelta {
		t.delta = t.maxDelta
	}
	t.delta32 = float32(t.delta)
	t.lastTime = now
}

// use this to signal that a rendering phase started
//
// combine this with renderCompleted and you can have
// the time spent rendering
func (t *Tick) beginRender() {
	t.lastRenderTime = glfw.GetTime()
}

func (t *Tick) renderCompleted() {
	t.lastRenderTime = t.lastRenderTime - glfw.GetTime()
}

// Display is where the interaction between player and game
// happens.
//
// A display usually renders a world that is composed by
// objects.
type Display struct {
	tick   *Tick
	window *glfw.Window
	// list of objects to render in the
	// next frame
	renderQueue []*Object
}

func Vsync(vsync bool) {
	if vsync {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}
}

func NewDisplay(width, height int, title string) (*Display, error) {
	err := ensureGlfwInit()
	if err != nil {
		return nil, err
	}
	d := &Display{
		tick: &Tick{},
	}

	// initialize the timers
	d.tick.tick()

	d.window, err = glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		return nil, err
	}
	useGlfwStats(func(s *glfwStats) {
		s.winCount++
	})
	return d, nil
}

func (d *Display) Close() {
	// see if we need to do anything else here....
	d.window = nil
	useGlfwStats(func(s *glfwStats) {
		s.winCount--
		if s.winCount == 0 {
			glfw.Terminate()
		}
	})
}

// Render will read the renderQueue and redraw them on the screen
// after this the renderQueue is empty.
func (d *Display) Render() {
	d.tick.beginRender()
	d.window.MakeContextCurrent()
	gl.Init()
	for _, obj := range d.renderQueue {
		obj.render(d.tick.delta32)
	}
	d.window.SwapBuffers()
	d.tick.renderCompleted()
	d.tick.tick()
}

// AcquireInput read input from the user devices
// (keyboards, mouse, gampepads).
//
// The input handling isn't done via triggers or
// channels but instead should be checked each frame.
//
// This function should be called at least one time each frame.
func (d *Display) AcquireInput() {
	glfw.PollEvents()
}

// ShouldClose returns true when the window
// should be closed. This is usually a result
// of the user pressing the close button.
func (d *Display) ShouldClose() bool {
	return d.window.ShouldClose()
}

func useGlfwStats(fn func(s *glfwStats)) {
	stats.Lock()
	defer stats.Unlock()
	fn(&stats)
}

func ensureGlfwInit() error {
	var err error
	useGlfwStats(func(glfwStats *glfwStats) {
		if glfwStats.started {
			err = nil
		}
		if glfw.Init() {
			err = nil
		} else {
			err = GlfwUnableToInit
		}
	})
	return err
}

type glfwStats struct {
	sync.Mutex
	winCount int
	started  bool
}

var (
	stats glfwStats
)
