package common

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"os"
)

// Log prints formatted, colored logs to the console
var Log consoleLogger

func init() {
	Log = consoleLogger{}
}

type consoleLogger struct {
	Debug  bool
	Silent bool
}

// Dump does a deep debug dump of a variable
func (log consoleLogger) Dump(a ...interface{}) {
	dumpper := spew.ConfigState{ContinueOnMethod: true}
	dump := dumpper.Sprintln(a)
	log.WriteDebug(dump)
}

// WriteDebug prints debug information to stdout
func (log consoleLogger) WriteDebug(format string, a ...interface{}) {
	if !log.Debug {
		return
	}

	color.Cyan(format, a...)
}

// WriteInfo prints text to stdout
func (log consoleLogger) WriteInfo(format string, a ...interface{}) {
	if log.Silent {
		return
	}

	color.White(format, a...)
}

// WriteWarning prints highlighted text to stdout
func (log consoleLogger) WriteWarning(format string, a ...interface{}) {
	if log.Silent {
		return
	}

	color.Yellow(format, a...)
}

// WriteError prints highlighted text and an error to stderr
func (log consoleLogger) WriteError(format string, err error, a ...interface{}) {
	color.Set(color.FgRed)
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	color.Unset()
}
