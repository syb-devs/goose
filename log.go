package goose

import (
	"os"

	"github.com/syb-devs/gotools/log"
)

// Log is the global app logger
var Log log.Logger

func init() {
	Log = log.New(os.Stdout)
}
