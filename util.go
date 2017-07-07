package meshview

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fogleman/fauxgl"
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

func boxForData(data []float32) fauxgl.Box {
	minx := data[0]
	maxx := data[0]
	miny := data[1]
	maxy := data[1]
	minz := data[2]
	maxz := data[2]
	for i := 0; i < len(data); i += 3 {
		x := data[i+0]
		y := data[i+1]
		z := data[i+2]
		if x < minx {
			minx = x
		}
		if x > maxx {
			maxx = x
		}
		if y < miny {
			miny = y
		}
		if y > maxy {
			maxy = y
		}
		if z < minz {
			minz = z
		}
		if z > maxz {
			maxz = z
		}
	}
	min := fauxgl.Vector{float64(minx), float64(miny), float64(minz)}
	max := fauxgl.Vector{float64(maxx), float64(maxy), float64(maxz)}
	return fauxgl.Box{min, max}
}
