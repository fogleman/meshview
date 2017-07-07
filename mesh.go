package meshview

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fogleman/fauxgl"
	"github.com/go-gl/gl/v2.1/gl"
)

func LoadMesh(path string) (*MeshData, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".stl":
		return LoadSTL(path)
	case ".obj":
		return LoadOBJ(path)
	}
	return nil, fmt.Errorf("unrecognized mesh extension: %s", ext)
}

type MeshData struct {
	Buffer []float32
	Box    fauxgl.Box
}

type Mesh struct {
	Transform    fauxgl.Matrix
	VertexBuffer uint32
	VertexCount  int32
}

func NewMesh(data *MeshData) *Mesh {
	// compute transform to scale and center mesh
	scale := fauxgl.V(2, 2, 2).Div(data.Box.Size()).MinComponent()
	transform := fauxgl.Identity()
	transform = transform.Translate(data.Box.Center().Negate())
	transform = transform.Scale(fauxgl.V(scale, scale, scale))

	// generate vbo
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(data.Buffer)*4, gl.Ptr(data.Buffer), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	// compute number of vertices
	count := int32(len(data.Buffer) / 3)

	return &Mesh{transform, vbo, count}
}

func (mesh *Mesh) Draw(positionAttrib uint32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, mesh.VertexBuffer)
	gl.EnableVertexAttribArray(positionAttrib)
	gl.VertexAttribPointer(positionAttrib, 3, gl.FLOAT, false, 12, gl.PtrOffset(0))
	gl.DrawArrays(gl.TRIANGLES, 0, mesh.VertexCount)
	gl.DisableVertexAttribArray(positionAttrib)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}
