package log

import (
	"fmt"
	"io"
)

type IOLogger struct {
	normW io.Writer
	errW  io.Writer
}

func NewIOLogger(out io.Writer, err io.Writer) *IOLogger {
	return &IOLogger{
		normW: out,
		errW:  err,
	}
}

func (l *IOLogger) Info(format string, a ...any) {
	fmt.Fprintf(l.normW, format, a...)
}

func (l *IOLogger) Error(format string, a ...any) {
	fmt.Fprintf(l.errW, format, a...)
}
