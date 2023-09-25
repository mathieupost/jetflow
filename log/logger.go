package log

import (
	"os"

	"github.com/rs/zerolog"
)

var logger = zerolog.New(os.Stdout)

const enabled = true

func Print(v ...interface{}) {
	if !enabled {
		return
	}
	logger.Print(v...)
}

func Println(v ...interface{}) {
	Print(v...)
}

func Fatalln(v ...interface{}) {
	Print(v...)
	os.Exit(1)
}

func Fatal(v ...interface{}) {
	Print(v...)
	os.Exit(1)
}
