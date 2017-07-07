package meshview

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func parseIndex(value string, count int) int {
	parsed, _ := strconv.ParseInt(value, 0, 0)
	n := int(parsed)
	if n < 0 {
		n += count
	}
	return n
}

func LoadOBJ(path string) (*MeshData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	count := 1
	lookup := make([]float32, 3, 1024)
	var data []float32
	var indexes []int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		keyword := fields[0]
		args := fields[1:]
		switch keyword {
		case "v":
			x, _ := strconv.ParseFloat(args[0], 32)
			y, _ := strconv.ParseFloat(args[1], 32)
			z, _ := strconv.ParseFloat(args[2], 32)
			v := []float32{float32(x), float32(y), float32(z)}
			lookup = append(lookup, v...)
			count++
		case "f":
			indexes = indexes[:0]
			for _, arg := range args {
				i := strings.Index(arg, "/")
				if i >= 0 {
					arg = arg[:i]
				}
				index := parseIndex(arg, count)
				indexes = append(indexes, index)
			}
			for i := 1; i < len(indexes)-1; i++ {
				i1 := indexes[0] * 3
				i2 := indexes[i] * 3
				i3 := indexes[i+1] * 3
				data = append(data, lookup[i1:i1+3]...)
				data = append(data, lookup[i2:i2+3]...)
				data = append(data, lookup[i3:i3+3]...)
			}
		}
	}

	box := boxForData(data)
	return &MeshData{data, box}, scanner.Err()
}
