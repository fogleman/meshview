package meshview

import (
	"github.com/fogleman/fauxgl"
	"github.com/go-gl/gl/v2.1/gl"
)

type MeshData struct {
	Buffer []float32
	Color  []uint16
	Box    fauxgl.Box
}

type Mesh struct {
	Transform      fauxgl.Matrix
	VertexCount    int32
	PositionBuffer uint32
	ColorBuffer    uint32
}

func NewMesh(data *MeshData) *Mesh {
	// compute transform to scale and center mesh
	scale := fauxgl.V(2, 2, 2).Div(data.Box.Size()).MinComponent()
	transform := fauxgl.Identity()
	transform = transform.Translate(data.Box.Center().Negate())
	transform = transform.Scale(fauxgl.V(scale, scale, scale))

	// generate vbo
	var position uint32
	gl.GenBuffers(1, &position)
	gl.BindBuffer(gl.ARRAY_BUFFER, position)
	if len(data.Buffer) > 0 {
		gl.BufferData(gl.ARRAY_BUFFER, len(data.Buffer)*4, gl.Ptr(data.Buffer), gl.STATIC_DRAW)
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	var color uint32
	gl.GenBuffers(1, &color)
	gl.BindBuffer(gl.ARRAY_BUFFER, color)
	if len(data.Color) > 0 {
		gl.BufferData(gl.ARRAY_BUFFER, len(data.Color)*2, gl.Ptr(data.Color), gl.STATIC_DRAW)
	} else {
		color := make([]uint16, len(data.Buffer)/3)
		gl.BufferData(gl.ARRAY_BUFFER, len(color)*2, gl.Ptr(color), gl.STATIC_DRAW)
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	// compute number of vertices
	count := int32(len(data.Buffer) / 3)

	return &Mesh{transform, count, position, color}
}

func (mesh *Mesh) Draw(positionAttrib, colorAttrib uint32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, mesh.PositionBuffer)
	gl.EnableVertexAttribArray(positionAttrib)
	gl.VertexAttribPointer(positionAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

	gl.BindBuffer(gl.ARRAY_BUFFER, mesh.ColorBuffer)
	gl.EnableVertexAttribArray(colorAttrib)
	gl.VertexAttribPointer(colorAttrib, 1, gl.UNSIGNED_SHORT, false, 0, gl.PtrOffset(0))

	gl.DrawArrays(gl.TRIANGLES, 0, mesh.VertexCount)

	gl.DisableVertexAttribArray(positionAttrib)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.DisableVertexAttribArray(colorAttrib)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

func (mesh *Mesh) Destroy() {
	gl.DeleteBuffers(1, &mesh.PositionBuffer)
}
