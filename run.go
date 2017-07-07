package meshview

import (
	"fmt"
	"runtime"
	"time"

	"github.com/fogleman/fauxgl"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var vertexShader = `
#version 120

uniform mat4 matrix;

attribute vec4 position;

varying vec3 ec_pos;

void main() {
	gl_Position = matrix * position;
	ec_pos = vec3(gl_Position);
}
`

var fragmentShader = `
#version 120

varying vec3 ec_pos;

const vec3 light_direction = normalize(vec3(1, -1, 1));
const vec3 object_color = vec3(52 / 255.0, 152 / 255.0, 219 / 255.0);

void main() {
	vec3 ec_normal = normalize(cross(dFdx(ec_pos), dFdy(ec_pos)));
	float diffuse = max(0, dot(ec_normal, light_direction)) * 0.9 + 0.1;
	vec3 color = object_color * diffuse;
	gl_FragColor = vec4(color, 1);
}
`

func init() {
	runtime.LockOSThread()
}

func loadMesh(path string) chan *MeshData {
	ch := make(chan *MeshData)
	go func() {
		start := time.Now()
		data, err := LoadMesh(path)
		if err != nil {
			panic(err)
		}
		fmt.Printf(
			"loaded %d triangles in %.3f seconds\n",
			len(data.Buffer)/9, time.Since(start).Seconds())
		ch <- data
		close(ch)
	}()
	return ch
}

func Run(path string) {
	start := time.Now()

	// load mesh in the background
	ch := loadMesh(path)

	// initialize glfw
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// create the window
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(640, 640, path, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	fmt.Printf("window shown at %.3f seconds\n", time.Since(start).Seconds())

	// initialize gl
	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.ClearColor(0.83, 0.85, 0.87, 1)

	// compile shaders
	program, err := compileProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)

	matrixUniform := uniformLocation(program, "matrix")
	positionAttrib := attribLocation(program, "position")

	// wait for mesh data to be loaded
	var data *MeshData
	for !window.ShouldClose() && data == nil {
		select {
		case data = <-ch:
		default:
			gl.Clear(gl.COLOR_BUFFER_BIT)
			window.SwapBuffers()
			glfw.PollEvents()
		}
	}
	if window.ShouldClose() {
		return
	}

	// create vbo and interactor
	mesh := NewMesh(data)
	// interactor := NewTurntable()
	interactor := NewArcball()
	BindInteractor(window, interactor)

	// render function
	render := func() {
		gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)
		matrix := getMatrix(window, interactor, mesh)
		setMatrix(matrixUniform, matrix)
		mesh.Draw(positionAttrib)
		window.SwapBuffers()
	}

	// render during resize
	window.SetFramebufferSizeCallback(func(window *glfw.Window, w, h int) {
		render()
	})

	fmt.Printf("first frame at %.3f seconds\n", time.Since(start).Seconds())

	// main loop
	for !window.ShouldClose() {
		render()
		glfw.PollEvents()
	}
}

func getMatrix(window *glfw.Window, interactor Interactor, mesh *Mesh) fauxgl.Matrix {
	matrix := mesh.Transform
	matrix = interactor.Matrix(window).Mul(matrix)
	return matrix
}
