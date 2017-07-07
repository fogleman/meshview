package meshview

import (
	"math"

	"github.com/fogleman/fauxgl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type Interactor interface {
	Matrix(window *glfw.Window) fauxgl.Matrix
	CursorPositionCallback(window *glfw.Window, x, y float64)
	MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)
	ScrollCallback(window *glfw.Window, dx, dy float64)
}

func BindInteractor(window *glfw.Window, interactor Interactor) {
	window.SetCursorPosCallback(glfw.CursorPosCallback(interactor.CursorPositionCallback))
	window.SetMouseButtonCallback(glfw.MouseButtonCallback(interactor.MouseButtonCallback))
	window.SetScrollCallback(glfw.ScrollCallback(interactor.ScrollCallback))
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

// func (t *Turntable) Matrix() fauxgl.Matrix {
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

type Arcball struct {
	Sensitivity float64
	Start       fauxgl.Vector
	Current     fauxgl.Vector
	Rotation    fauxgl.Matrix
	Translation fauxgl.Vector
	Scroll      float64
	Rotate      bool
	Pan         bool
}

func NewArcball() Interactor {
	a := Arcball{}
	a.Sensitivity = 6
	a.Rotation = fauxgl.Identity()
	return &a
}

func (a *Arcball) CursorPositionCallback(window *glfw.Window, x, y float64) {
	if a.Rotate {
		a.Current = arcballVector(window)
	}
	if a.Pan {
		a.Current = screenPosition(window)
	}
}

func (a *Arcball) MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		if action == glfw.Press {
			if mods == 0 {
				v := arcballVector(window)
				a.Start = v
				a.Current = v
				a.Rotate = true
			} else {
				v := screenPosition(window)
				a.Start = v
				a.Current = v
				a.Pan = true
			}
		} else if action == glfw.Release {
			if a.Rotate {
				m := arcballRotate(a.Start, a.Current, a.Sensitivity)
				a.Rotation = m.Mul(a.Rotation)
				a.Rotate = false
			}
			if a.Pan {
				d := a.Current.Sub(a.Start)
				a.Translation = a.Translation.Add(d)
				a.Pan = false
			}
		}
	}
}

func (a *Arcball) ScrollCallback(window *glfw.Window, dx, dy float64) {
	a.Scroll += dy
}

func (a *Arcball) Matrix(window *glfw.Window) fauxgl.Matrix {
	w, h := window.GetFramebufferSize()
	aspect := float64(w) / float64(h)
	r := a.Rotation
	if a.Rotate {
		r = arcballRotate(a.Start, a.Current, a.Sensitivity).Mul(r)
	}
	t := a.Translation
	if a.Pan {
		t = t.Add(a.Current.Sub(a.Start))
	}
	s := math.Pow(0.98, a.Scroll)
	m := fauxgl.Identity()
	m = m.Scale(fauxgl.V(s, s, s))
	m = r.Mul(m)
	m = m.LookAt(fauxgl.V(0, -5, 0), fauxgl.V(0, 0, 0), fauxgl.V(0, 0, 1))
	m = m.Perspective(30, aspect, 1, 10)
	m = m.Translate(t)
	return m
}

func screenPosition(window *glfw.Window) fauxgl.Vector {
	x, y := window.GetCursorPos()
	w, h := window.GetSize()
	x = (x/float64(w))*2 - 1
	y = (y/float64(h))*2 - 1
	return fauxgl.Vector{x, -y, 0}
}

func arcballVector(window *glfw.Window) fauxgl.Vector {
	x, y := window.GetCursorPos()
	w, h := window.GetSize()
	x = (x/float64(w))*2 - 1
	y = (y/float64(h))*2 - 1
	x /= 4
	y /= 4
	x = -x
	q := x*x + y*y
	if q <= 1 {
		z := math.Sqrt(1 - q)
		return fauxgl.Vector{x, z, y}
	} else {
		return fauxgl.Vector{x, 0, y}.Normalize()
	}
}

func arcballRotate(a, b fauxgl.Vector, sensitivity float64) fauxgl.Matrix {
	const eps = 1e-9
	dot := b.Dot(a)
	if math.Abs(dot) < eps || math.Abs(dot-1) < eps {
		return fauxgl.Identity()
	} else if math.Abs(dot+1) < eps {
		return fauxgl.Rotate(a.Perpendicular(), math.Pi*sensitivity)
	} else {
		angle := math.Acos(dot)
		v := b.Cross(a).Normalize()
		return fauxgl.Rotate(v, angle*sensitivity)
	}
}
