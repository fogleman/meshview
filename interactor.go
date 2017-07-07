package meshview

import (
	"math"

	"github.com/fogleman/fauxgl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type Interactor interface {
	Matrix() fauxgl.Matrix
	CursorPositionCallback(window *glfw.Window, x, y float64)
	MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)
	ScrollCallback(window *glfw.Window, dx, dy float64)
}

func BindInteractor(window *glfw.Window, interactor Interactor) {
	window.SetCursorPosCallback(glfw.CursorPosCallback(interactor.CursorPositionCallback))
	window.SetMouseButtonCallback(glfw.MouseButtonCallback(interactor.MouseButtonCallback))
	window.SetScrollCallback(glfw.ScrollCallback(interactor.ScrollCallback))
}

type Turntable struct {
	Sensitivity float64
	Dx, Dy      float64
	Px, Py      float64
	Scroll      float64
	Down        bool
}

func NewTurntable() Interactor {
	t := Turntable{}
	t.Sensitivity = 0.5
	return &t
}

func (t *Turntable) CursorPositionCallback(window *glfw.Window, x, y float64) {
	if t.Down {
		t.Dx += x - t.Px
		t.Dy += y - t.Py
		t.Px = x
		t.Py = y
		t.Dy = math.Max(t.Dy, -90/t.Sensitivity)
		t.Dy = math.Min(t.Dy, 90/t.Sensitivity)
	}
}

func (t *Turntable) MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		if action == glfw.Press {
			t.Down = true
			t.Px, t.Py = window.GetCursorPos()
		} else if action == glfw.Release {
			t.Down = false
		}
	}
}

func (t *Turntable) ScrollCallback(window *glfw.Window, dx, dy float64) {
	t.Scroll += dy
}

func (t *Turntable) Matrix() fauxgl.Matrix {
	s := math.Pow(0.98, t.Scroll)
	a1 := fauxgl.Radians(-t.Dx * t.Sensitivity)
	a2 := fauxgl.Radians(-t.Dy * t.Sensitivity)
	m := fauxgl.Identity()
	m = m.Scale(fauxgl.V(s, s, s))
	m = m.Rotate(fauxgl.V(math.Cos(a1), math.Sin(a1), 0), a2)
	m = m.Rotate(fauxgl.V(0, 0, 1), a1)
	return m
}
