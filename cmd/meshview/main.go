package main

import (
	"os"

	"github.com/fogleman/meshview"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		meshview.Run(args)
	} else {
		meshview.Run(nil)
	}
}
