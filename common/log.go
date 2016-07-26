package common

import (
	"fmt"
	"github.com/fatih/color"
	"os"
)

var Log consoleLogger

func init() {
	Log = consoleLogger{}
}

type consoleLogger struct {
	Debug  bool
	Silent bool
}

func (log consoleLogger) WriteDebug(format string, a ...interface{}) {
	if !log.Debug {
		return
	}

	color.Cyan(format, a...)
}

func (log consoleLogger) WriteInfo(format string, a ...interface{}) {
	if log.Silent {
		return
	}

	color.White(format, a...)
}

func (log consoleLogger) WriteWarning(format string, a ...interface{}) {
	if log.Silent {
		return
	}

	color.Yellow(format, a...)
}

func (log consoleLogger) WriteError(format string, err error, a ...interface{}) {
	color.Set(color.FgRed)
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	color.Unset()
}
