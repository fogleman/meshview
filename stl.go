package meshview

import (
	"bufio"
	"encoding/binary"
	"io"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func LoadSTL(path string) (*MeshData, error) {
	// open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// get file size
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// read header, get expected binary size
	type STLHeader struct {
		_     [80]uint8
		Count uint32
	}
	header := STLHeader{}
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	expectedSize := int64(header.Count)*50 + 84

	// parse ascii or binary stl
	if info.Size() == expectedSize {
		return loadSTLB(file, int(header.Count))
	} else {
		// rewind to start of file
		_, err = file.Seek(0, 0)
		if err != nil {
			return nil, err
		}
		return loadSTLA(file)
	}
}

func loadSTLA(file *os.File) (*MeshData, error) {
	var data []float32
	var x1, y1, z1, x2, y2, z2, x3, y3, z3 float32
	i := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) < 12 || line[0] != 'v' {
			continue
		}
		fields := strings.Fields(line[7:])
		if len(fields) != 3 {
			continue
		}
		x, _ := strconv.ParseFloat(fields[0], 32)
		y, _ := strconv.ParseFloat(fields[1], 32)
		z, _ := strconv.ParseFloat(fields[2], 32)
		switch i % 3 {
		case 0:
			x1 = float32(x)
			y1 = float32(y)
			z1 = float32(z)
		case 1:
			x2 = float32(x)
			y2 = float32(y)
			z2 = float32(z)
		case 2:
			x3 = float32(x)
			y3 = float32(y)
			z3 = float32(z)
			data = append(data, []float32{x1, y1, z1, x2, y2, z2, x3, y3, z3}...)
		}
		i++
	}
	box := boxForData(data)
	return &MeshData{data, box}, scanner.Err()
}

func makeFloat(b []byte) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(b))
}

func loadSTLB(file *os.File, count int) (*MeshData, error) {
	buf := make([]byte, count*50)
	_, err := io.ReadFull(file, buf)
	if err != nil {
		return nil, err
	}

	data := make([]float32, count*9)
	wn := runtime.NumCPU() - 1
	if wn < 1 {
		wn = 1
	}
	var wg sync.WaitGroup
	for wi := 0; wi < wn; wi++ {
		wg.Add(1)
		go func(wi int) {
			n := count / wn
			if count%wn > 0 {
				n++
			}
			i0 := n * wi
			i1 := i0 + n
			if i1 > count {
				i1 = count
			}
			for i := i0; i < i1; i++ {
				j := i * 9
				b := buf[i*50+12:]
				data[j+0] = makeFloat(b[0:])
				data[j+1] = makeFloat(b[4:])
				data[j+2] = makeFloat(b[8:])
				data[j+3] = makeFloat(b[12:])
				data[j+4] = makeFloat(b[16:])
				data[j+5] = makeFloat(b[20:])
				data[j+6] = makeFloat(b[24:])
				data[j+7] = makeFloat(b[28:])
				data[j+8] = makeFloat(b[32:])
			}
			wg.Done()
		}(wi)
	}
	wg.Wait()

	box := boxForData(data)
	return &MeshData{data, box}, nil
}
