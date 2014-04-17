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
	tick      *Tick
	window    *glfw.Window
	lastAtlas *Atlas
	// list of objects to render in the
	// next frame
	renderQueue cmdQueue
	id          int
}

type cmdQueue struct {
	sync.RWMutex
	cmds []DrawCmd
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
	d.renderQueue.cmds = make([]DrawCmd, 0, 0)

	// initialize the timers
	d.tick.tick()

	d.window, err = glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		return nil, err
	}
	useGlfwStats(func(s *glfwStats) {
		s.winCount++
		d.id = s.winCount
	})
	return d, nil
}

func (d *Display) Close() {
	// see if we need to do anything else here....
	d.window.Destroy()
	d.window = nil
	useGlfwStats(func(s *glfwStats) {
		s.winCount--
		if s.winCount == 0 {
			glfw.Terminate()
		}
	})
}

// SendDraw receives a DrawCmd and enqueue for later rendering.
//
// The order of rendering is determinated by the Render function and it
// might or might not be the order of SendDraw.
//
// This method is safe to be used by multiple threads.
func (d *Display) SendDraw(cmd DrawCmd) {
	d.renderQueue.Lock()
	defer d.renderQueue.Unlock()
	d.renderQueue.cmds = append(d.renderQueue.cmds, cmd)
}

// Render will read the renderQueue and redraw them on the screen
// after this the renderQueue is empty.
func (d *Display) Render() {
	d.tick.beginRender()
	d.window.MakeContextCurrent()
	gl.Init()
	d.renderQueue.RLock()
	defer d.renderQueue.RUnlock()
	if len(d.renderQueue.cmds) > 0 {
		for _, obj := range d.renderQueue.cmds {
			err := obj.Render(d)
			if err != nil {
				d.logErr(err)
			}
		}
		d.renderQueue.cmds = d.renderQueue.cmds[:0]
		d.window.SwapBuffers()
	}
	d.tick.renderCompleted()
	d.tick.tick()
}

func (d *Display) logErr(err error) {
	log.Printf("[error]-[%v] %v", d.id, err)
}

func (d *Display) logInfo(info string) {
	log.Printf("[info]-[%v] %v", d.id, info)
}

func (d *Display) bindAtlas(a *Atlas) error {
	if d.lastAtlas == a {
		return nil
	}
	if err := d.discardAtlas(); err != nil {
		return err
	}
	d.lastAtlas = a
	return d.lastAtlas.bind()
}

func (d *Display) discardAtlas() error {
	if d.lastAtlas == nil {
		return nil
	}
	// unbind and delete from video memory, just to be safe
	return d.lastAtlas.unbind(true)
}

// AcquireInput read input from the user devices
// (keyboards, mouse, gampepads).
//
// The input handling isn't done via triggers or
// channels but instead should be checked each frame.
//
// This function should be called at least one time each frame.
func AcquireInput() {
	glfw.PollEvents()
}

// ShouldClose returns true when the window
// should be closed. This is usually a result
// of the user pressing the close button.
func (d *Display) ShouldClose() bool {
	return d.window.ShouldClose()
}

type DisplayList []*Display

func NewDisplayList() *DisplayList {
	dl := make(DisplayList, 0)
	return &dl
}

func (dl *DisplayList) Push(d ...*Display) {
	for _, nd := range d {
		*dl = append(*dl, nd)
	}
}

func (dl *DisplayList) Remove(d *Display) {
	slice := *dl
	removedIdx := int(-1)
	for i, v := range slice {
		if v == d {
			slice[i] = nil
			removedIdx = i
		}
	}
	if removedIdx >= 0 {
		dl.shrinkAt(removedIdx)
	}
}

// ShouldClose returns true when every window
// on this display is should be closed.
//
// Windows marked to be closed are removed from
// this list
func (dl *DisplayList) ShouldClose() bool {
	// worst case, everybody want's to
	// close
	toRemove := make([]*Display, 0, len(*dl))
	for _, v := range *dl {
		if v.ShouldClose() {
			toRemove = append(toRemove, v)
			v.Close()
		}
	}
	for _, v := range toRemove {
		dl.Remove(v)
	}
	return len(*dl) == 0
}

func (dl *DisplayList) Render() {
	for _, d := range *dl {
		d.Render()
	}
}

func (dl *DisplayList) shrinkAt(idx int) {
	slice := *dl
	slice[idx] = nil
	for i := idx; i < len(slice)-1; i++ {
		slice[i] = slice[i+1]
	}
	slice = slice[:len(slice)-1]
	*dl = slice
	if len(slice) == 0 {
		// release the space from
		// the backing array
		*dl = make([]*Display, 0)
	}
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
