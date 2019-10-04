package meshview

import (
	"math"

	"github.com/fogleman/fauxgl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

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
	a.Sensitivity = 20
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

func (a *Arcball) KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press && mods == 0 {
		if key >= 49 && key <= 55 {
			a.Translation = fauxgl.Vector{}
			a.Scroll = 0
		}
		switch key {
		case 49: //1
			a.Rotation = fauxgl.Identity()
		case 50:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(0, 0, 1), math.Pi/2)
		case 51:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(0, 0, 1), math.Pi)
		case 52:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(0, 0, 1), -math.Pi/2)
		case 53:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(1, 0, 0), math.Pi/2)
		case 54:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(1, 0, 0), -math.Pi/2)
		case 55:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(1, 1, 0).Normalize(), -math.Pi/4).Rotate(fauxgl.V(0, 0, 1), math.Pi/4)
		case 88: //X
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
	m = m.Translate(t)
	m = m.LookAt(fauxgl.V(0, -3, 0), fauxgl.V(0, 0, 0), fauxgl.V(0, 0, 1))
	m = m.Perspective(50, aspect, 0.1, 100)
	return m
}

func screenPosition(window *glfw.Window) fauxgl.Vector {
	x, y := window.GetCursorPos()
	w, h := window.GetSize()
	x = (x/float64(w))*2 - 1
	y = (y/float64(h))*2 - 1
	return fauxgl.Vector{x, 0, -y}
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
