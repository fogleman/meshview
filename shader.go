package meshview

import (
	"fmt"
	"strings"

	"github.com/fogleman/fauxgl"
	"github.com/go-gl/gl/v2.1/gl"
)

func setMatrix(location int32, m fauxgl.Matrix) {
	data := [16]float32{
		float32(m.X00), float32(m.X01), float32(m.X02), float32(m.X03),
		float32(m.X10), float32(m.X11), float32(m.X12), float32(m.X13),
		float32(m.X20), float32(m.X21), float32(m.X22), float32(m.X23),
		float32(m.X30), float32(m.X31), float32(m.X32), float32(m.X33),
	}
	gl.UniformMatrix4fv(location, 1, true, &data[0])
}

func uniformLocation(program uint32, name string) int32 {
	return gl.GetUniformLocation(program, gl.Str(name+"\x00"))
}

func attribLocation(program uint32, name string) uint32 {
	return uint32(gl.GetAttribLocation(program, gl.Str(name+"\x00")))
}

func compileProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf(log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)
	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf(log)
	}

	return shader, nil
}
