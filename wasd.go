package meshview

import (
	"math"
	"time"

	"github.com/fogleman/fauxgl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type WASD struct {
	sensitivity float64
	invert      bool
	discard     bool
	previous    time.Time
	position    fauxgl.Vector
	mx, my      float64
	rx, ry      float64
}

func NewWASD(window *glfw.Window) Interactor {
	wasd := WASD{}
	wasd.position = fauxgl.V(0, -3, 0)
	wasd.sensitivity = 2.5
	if window != nil {
		wasd.setExclusive(window, true)
	}
	return &wasd
}

func (wasd *WASD) isExclusive(window *glfw.Window) bool {
	return window.GetInputMode(glfw.CursorMode) == glfw.CursorDisabled
}

func (wasd *WASD) setExclusive(window *glfw.Window, exclusive bool) {
	if exclusive {
		wasd.previous = time.Now()
		wasd.discard = true
		window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
}

func (wasd *WASD) sightVector() fauxgl.Vector {
	v := fauxgl.V(0, 1, 0)
	v = fauxgl.Rotate(fauxgl.V(1, 0, 0), wasd.ry).MulDirection(v)
	v = fauxgl.Rotate(fauxgl.V(0, 0, 1), wasd.rx).MulDirection(v)
	return v
}

func (wasd *WASD) motionVector(sx, sy, sz int, dt float64) fauxgl.Vector {
	up := fauxgl.V(0, 0, 1)
	sv := wasd.sightVector()
	pv := sv.Cross(up)
	var v fauxgl.Vector
	v = v.Add(sv.MulScalar(float64(-sy)))
	v = v.Add(pv.MulScalar(float64(sx)))
	v = v.Add(up.MulScalar(float64(sz)))
	if v.Length() > 0 {
		v = v.Normalize()
	}
	return v
}

func (wasd *WASD) strafe(window *glfw.Window) (int, int, int) {
	var sx, sy, sz int
	if window.GetKey(glfw.KeyA) == glfw.Press {
		sx--
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		sx++
	}
	if window.GetKey(glfw.KeyW) == glfw.Press {
		sy--
	}
	if window.GetKey(glfw.KeyS) == glfw.Press {
		sy++
	}
	if window.GetKey(glfw.KeySpace) == glfw.Press {
		sz++
	}
	return sx, sy, sz
}

func (wasd *WASD) updatePosition(window *glfw.Window, dt float64) {
	sx, sy, sz := wasd.strafe(window)
	mv := wasd.motionVector(sx, sy, sz, dt)
	wasd.position = wasd.position.Add(mv.MulScalar(dt * 1))
}

func (wasd *WASD) CursorPositionCallback(window *glfw.Window, x, y float64) {
	if !wasd.isExclusive(window) {
		return
	}
	if wasd.discard {
		wasd.mx = x
		wasd.my = y
		wasd.discard = false
		return
	}

	m := wasd.sensitivity / 1000.0
	wasd.rx += (x - wasd.mx) * m
	if wasd.invert {
		wasd.ry -= (y - wasd.my) * m
	} else {
		wasd.ry += (y - wasd.my) * m
	}
	if wasd.rx < 0 {
		wasd.rx += 2 * math.Pi
	}
	if wasd.rx >= 2*math.Pi {
		wasd.rx -= 2 * math.Pi
	}
	wasd.ry = math.Max(wasd.ry, -math.Pi/2)
	wasd.ry = math.Min(wasd.ry, math.Pi/2)
	wasd.mx = x
	wasd.my = y
}

func (wasd *WASD) MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if !wasd.isExclusive(window) {
		if button == glfw.MouseButton1 && action == glfw.Press {
			wasd.setExclusive(window, true)
		}
	}
}

func (wasd *WASD) KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if wasd.isExclusive(window) {
		if key == glfw.KeyEscape {
			wasd.setExclusive(window, false)
		}
	}
}

func (wasd *WASD) ScrollCallback(window *glfw.Window, dx, dy float64) {
}

func (wasd *WASD) Matrix(window *glfw.Window) fauxgl.Matrix {
	now := time.Now()
	dt := now.Sub(wasd.previous).Seconds()
	wasd.previous = now

	wasd.updatePosition(window, dt)

	w, h := window.GetFramebufferSize()
	aspect := float64(w) / float64(h)
	eye := wasd.position
	center := eye.Add(wasd.sightVector())

	m := fauxgl.Identity()
	m = m.LookAt(eye, center, fauxgl.V(0, 0, 1))
	m = m.Perspective(50, aspect, 0.01, 100)
	return m
}
