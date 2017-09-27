package lib

import (
	"log"
	"os"
	"path/filepath"
)

var Logger *log.Logger

func init() {
	prefix := "[" + filepath.Base(os.Args[0]) + "] "
	Logger = log.New(os.Stdout, prefix, log.Lshortfile)
}
