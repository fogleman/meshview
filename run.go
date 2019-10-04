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

const vec3 light_direction = normalize(vec3(1, -1.5, 1));
const vec3 object_color = vec3(0x5b / 255.0, 0xac / 255.0, 0xe3 / 255.0);

void main() {
	vec3 ec_normal = normalize(cross(dFdx(ec_pos), dFdy(ec_pos)));
	float diffuse = max(0, dot(ec_normal, light_direction)) * 0.9 + 0.15;
	vec3 color = object_color * diffuse;
	gl_FragColor = vec4(color, 1);
}
`

func init() {
	runtime.LockOSThread()
}

func loadMesh(path string, ch chan *MeshData) {
	go func() {
		start := time.Now()
		data, err := LoadMesh(path)
		if err != nil {
			return // TODO: display an error
		}
		fmt.Printf(
			"loaded %d triangles in %.3f seconds\n",
			len(data.Buffer)/9, time.Since(start).Seconds())
		ch <- data
	}()
}

func Run(path string) {
	start := time.Now()

	// load mesh in the background
	ch := make(chan *MeshData)
	loadMesh(path, ch)

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
	gl.ClearColor(float32(0xd4)/255, float32(0xd9)/255, float32(0xde)/255, 1)

	// compile shaders
	program, err := compileProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)

	matrixUniform := uniformLocation(program, "matrix")
	positionAttrib := attribLocation(program, "position")

	var mesh *Mesh

	// create interactor
	interactor := NewSwitchableInteractor([]Interactor{
		NewArcball(),
		NewWASD(nil),
	})
	BindInteractor(window, interactor)

	// render function
	render := func() {
		gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)
		if mesh != nil {
			matrix := getMatrix(window, interactor, mesh)
			setMatrix(matrixUniform, matrix)
			mesh.Draw(positionAttrib)
		}
		window.SwapBuffers()
	}

	// render during resize
	window.SetFramebufferSizeCallback(func(window *glfw.Window, w, h int) {
		render()
	})

	// handle drop events
	window.SetDropCallback(func(window *glfw.Window, filenames []string) {
		loadMesh(filenames[0], ch)
		window.SetTitle(filenames[0])
	})

	// main loop
	for !window.ShouldClose() {
		select {
		case data := <-ch:
			if mesh != nil {
				mesh.Destroy()
			}
			mesh = NewMesh(data)
			fmt.Printf("first frame at %.3f seconds\n", time.Since(start).Seconds())
		default:
		}
		render()
		glfw.PollEvents()
	}
}

func getMatrix(window *glfw.Window, interactor Interactor, mesh *Mesh) fauxgl.Matrix {
	matrix := mesh.Transform
	matrix = interactor.Matrix(window).Mul(matrix)
	return matrix
}
