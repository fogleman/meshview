package main

import (
	"log"
	"os"

	"github.com/fogleman/meshview"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatalln("Usage: meshview input.stl")
	}
	meshview.Run(args[0])
}
