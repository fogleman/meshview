package meshview

import (
	"github.com/fogleman/fauxgl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type Interactor interface {
	Matrix(window *glfw.Window) fauxgl.Matrix
	CursorPositionCallback(window *glfw.Window, x, y float64)
	MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)
	KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)
	ScrollCallback(window *glfw.Window, dx, dy float64)
}

func BindInteractor(window *glfw.Window, interactor Interactor) {
	window.SetCursorPosCallback(glfw.CursorPosCallback(interactor.CursorPositionCallback))
	window.SetMouseButtonCallback(glfw.MouseButtonCallback(interactor.MouseButtonCallback))
	window.SetKeyCallback(glfw.KeyCallback(interactor.KeyCallback))
	window.SetScrollCallback(glfw.ScrollCallback(interactor.ScrollCallback))
}

type SwitchableInteractor struct {
	Interactors []Interactor
	Index       int
}

func NewSwitchableInteractor(interactors []Interactor) *SwitchableInteractor {
	return &SwitchableInteractor{interactors, 0}
}

func (si *SwitchableInteractor) Switch() {
	si.Index = (si.Index + 1) % len(si.Interactors)
}

func (si *SwitchableInteractor) Matrix(window *glfw.Window) fauxgl.Matrix {
	return si.Interactors[si.Index].Matrix(window)
}

func (si *SwitchableInteractor) CursorPositionCallback(window *glfw.Window, x, y float64) {
	si.Interactors[si.Index].CursorPositionCallback(window, x, y)
}

func (si *SwitchableInteractor) MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	si.Interactors[si.Index].MouseButtonCallback(window, button, action, mods)
}

func (si *SwitchableInteractor) KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyTab && action == glfw.Press {
		si.Switch()
	}
	si.Interactors[si.Index].KeyCallback(window, key, scancode, action, mods)
}

func (si *SwitchableInteractor) ScrollCallback(window *glfw.Window, dx, dy float64) {
	si.Interactors[si.Index].ScrollCallback(window, dx, dy)
}

// Turntable

// type Turntable struct {
// 	Sensitivity float64
// 	Dx, Dy      float64
// 	Px, Py      float64
// 	Scroll      float64
// 	Rotate      bool
// }

// func NewTurntable() Interactor {
// 	t := Turntable{}
// 	t.Sensitivity = 0.5
// 	return &t
// }

// func (t *Turntable) CursorPositionCallback(window *glfw.Window, x, y float64) {
// 	if t.Rotate {
// 		t.Dx += x - t.Px
// 		t.Dy += y - t.Py
// 		t.Px = x
// 		t.Py = y
// 		t.Dy = math.Max(t.Dy, -90/t.Sensitivity)
// 		t.Dy = math.Min(t.Dy, 90/t.Sensitivity)
// 	}
// }

// func (t *Turntable) MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
// 	if button == glfw.MouseButton1 {
// 		if action == glfw.Press {
// 			t.Rotate = true
// 			t.Px, t.Py = window.GetCursorPos()
// 		} else if action == glfw.Release {
// 			t.Rotate = false
// 		}
// 	}
// }

// func (t *Turntable) ScrollCallback(window *glfw.Window, dx, dy float64) {
// 	t.Scroll += dy
// }

// func (t *Turntable) KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
// }

// func (t *Turntable) Matrix(window *glfw.Window) fauxgl.Matrix {
// 	s := math.Pow(0.98, t.Scroll)
// 	a1 := fauxgl.Radians(-t.Dx * t.Sensitivity)
// 	a2 := fauxgl.Radians(-t.Dy * t.Sensitivity)
// 	m := fauxgl.Identity()
// 	m = m.Scale(fauxgl.V(s, s, s))
// 	m = m.Rotate(fauxgl.V(math.Cos(a1), math.Sin(a1), 0), a2)
// 	m = m.Rotate(fauxgl.V(0, 0, 1), a1)
// 	return m
// }

// func (t *Turntable) Translation() fauxgl.Vector {
// 	return fauxgl.Vector{}
// }

// Arcball
