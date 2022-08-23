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
	fauxgl.HexColor("#e41a1c"),
	fauxgl.HexColor("#377eb8"),
	fauxgl.HexColor("#4daf4a"),
	fauxgl.HexColor("#984ea3"),
	fauxgl.HexColor("#ff7f00"),
	fauxgl.HexColor("#ffff33"),
	fauxgl.HexColor("#a65628"),
	fauxgl.HexColor("#f781bf"),
}

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

uniform vec3 object_color;

varying vec3 ec_pos;

const vec3 light_direction = normalize(vec3(1, -1.5, 1));

void main() {
	vec3 ec_normal = normalize(cross(dFdx(ec_pos), dFdy(ec_pos)));
	float diffuse = max(0, dot(ec_normal, light_direction)) * 0.9 + 0.25;
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

func Run(paths []string) {
	start := time.Now()

	ch := make(chan *MeshData)

	// load mesh in the background
	for _, path := range paths {
		loadMesh(path, ch)
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

	matrixUniform := uniformLocation(program, "matrix")
	positionAttrib := attribLocation(program, "position")
	objectColorUniform := uniformLocation(program, "object_color")

	var meshes []*Mesh

	// create interactor
	interactor := NewSwitchableInteractor([]Interactor{
		NewArcball(),
		NewWASD(nil),
	})
	BindInteractor(window, interactor)

	// render function
	render := func() {
		gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)
		if len(meshes) > 0 {
			matrix := getMatrix(window, interactor, meshes[0])
			setMatrix(matrixUniform, matrix)
			for i, mesh := range meshes {
				c := objectColors[i%len(objectColors)]
				r, g, b := float32(c.R), float32(c.G), float32(c.B)
				gl.Uniform3f(objectColorUniform, r, g, b)
				mesh.Draw(positionAttrib)
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
			loadMesh(path, ch)
		}
	})

	// main loop
	for !window.ShouldClose() {
		select {
		case data := <-ch:
			meshes = append(meshes, NewMesh(data))
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
