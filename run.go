package meshview

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/fogleman/fauxgl"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var objectColors = []fauxgl.Color{
	fauxgl.HexColor("#377eb8"),
	fauxgl.HexColor("#e41a1c"),
	fauxgl.HexColor("#4daf4a"),
	fauxgl.HexColor("#984ea3"),
	fauxgl.HexColor("#ff7f00"),
	fauxgl.HexColor("#ffff33"),
	fauxgl.HexColor("#a65628"),
	fauxgl.HexColor("#f781bf"),
}

var vertexShader = `
#version 120

uniform mat4 viewMatrix;
uniform mat4 perspectiveMatrix;
uniform vec3 object_color;

attribute vec4 position;
attribute float color;

varying vec3 ec_pos;
varying vec3 ec_color;

void main() {
	gl_Position = perspectiveMatrix * position;
	ec_pos = vec3(viewMatrix * position);
	ec_color = object_color;
	if (color != 0) {
		float r = mod(floor(color / 1024), 32) / 31;
		float g = mod(floor(color / 32), 32) / 31;
		float b = mod(floor(color / 1), 32) / 31;
		ec_color = vec3(r, g, b);
	}
}
`

var fragmentShader = `
#version 120

varying vec3 ec_pos;
varying vec3 ec_color;

const vec3 light_direction = normalize(vec3(0.5, 0.5, 1));

void main() {
	vec3 ec_normal = normalize(cross(dFdx(ec_pos), dFdy(ec_pos)));
	float diffuse = max(0, dot(ec_normal, light_direction)) * 0.9 + 0.25;
	vec3 color = ec_color * diffuse;
	gl_FragColor = vec4(color, 1);
}
`

func init() {
	runtime.LockOSThread()
}

type LoadedMesh struct {
	Index int
	Data  *MeshData
}

func loadMesh(index int, path string, ch chan LoadedMesh) {
	go func() {
		start := time.Now()
		data, err := LoadMesh(path)
		if err != nil {
			return // TODO: display an error
		}
		fmt.Printf(
			"loaded %d triangles in %.3f seconds\n",
			len(data.Buffer)/9, time.Since(start).Seconds())
		ch <- LoadedMesh{index, data}
	}()
}

func Run(paths []string) {
	start := time.Now()

	ch := make(chan LoadedMesh)

	// load mesh in the background
	meshes := make([]*Mesh, len(paths))
	for i, path := range paths {
		loadMesh(i, path, ch)
	}

	// initialize glfw
	if err := glfw.Init(); err != nil {
		log.Fatal(err)
	}
	defer glfw.Terminate()

	// create the window
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(640, 640, "Mesh Viewer", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	window.MakeContextCurrent()

	fmt.Printf("window shown at %.3f seconds\n", time.Since(start).Seconds())

	// initialize gl
	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.ClearColor(float32(0xd4)/255, float32(0xd9)/255, float32(0xde)/255, 1)

	// compile shaders
	program, err := compileProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Fatal(err)
	}
	gl.UseProgram(program)

	viewMatrixUniform := uniformLocation(program, "viewMatrix")
	perspectiveMatrixUniform := uniformLocation(program, "perspectiveMatrix")
	positionAttrib := attribLocation(program, "position")
	colorAttrib := attribLocation(program, "color")
	objectColorUniform := uniformLocation(program, "object_color")

	// create interactor
	interactor := NewSwitchableInteractor([]Interactor{
		NewArcball(),
		NewWASD(nil),
	})
	BindInteractor(window, interactor)

	// render function
	render := func() {
		gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)
		if len(meshes) > 0 && meshes[0] != nil {
			w, h := window.GetFramebufferSize()
			aspect := float64(w) / float64(h)
			viewMatrix := getMatrix(window, interactor, meshes[0])
			var perspectiveMatrix fauxgl.Matrix
			if interactor.Orthographic() {
				const s = 1.2
				perspectiveMatrix = viewMatrix.Orthographic(-s*aspect, s*aspect, -s, s, -100, 100)
			} else {
				perspectiveMatrix = viewMatrix.Perspective(50, aspect, 0.01, 1000)
			}
			setMatrix(viewMatrixUniform, viewMatrix)
			setMatrix(perspectiveMatrixUniform, perspectiveMatrix)
			for i, mesh := range meshes {
				if mesh == nil {
					continue
				}
				if !interactor.Visible(i) {
					continue
				}
				c := objectColors[i%len(objectColors)]
				r, g, b := float32(c.R), float32(c.G), float32(c.B)
				gl.Uniform3f(objectColorUniform, r, g, b)
				mesh.Draw(positionAttrib, colorAttrib)
			}
		}
		window.SwapBuffers()
	}

	// render during resize
	window.SetFramebufferSizeCallback(func(window *glfw.Window, w, h int) {
		render()
	})

	// handle drop events
	window.SetDropCallback(func(window *glfw.Window, filenames []string) {
		for _, path := range filenames {
			i := len(meshes)
			meshes = append(meshes, nil)
			loadMesh(i, path, ch)
		}
	})

	// main loop
	for !window.ShouldClose() {
		select {
		case loadedMesh := <-ch:
			meshes[loadedMesh.Index] = NewMesh(loadedMesh.Data)
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
