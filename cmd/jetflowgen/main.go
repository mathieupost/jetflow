package main

import (
	"os"

	"github.com/mathieupost/jetflow/generate"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	generate.ParsePackage(cwd)
}
