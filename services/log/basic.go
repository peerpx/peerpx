package log

import (
	"fmt"
	"io"
	logStd "log"
)

type basicLogger struct {
	*logStd.Logger
}

func InitBasicLogger(output io.Writer) error {
	l = &basicLogger{logStd.New(output, "peerpx - ", logStd.LstdFlags)}
	return nil
}

func (l *basicLogger) info(v ...interface{}) {
	l.Logger.Print("- info: ", fmt.Sprintln(v...))
}

func (l *basicLogger) infof(format string, v ...interface{}) {
	format = fmt.Sprintf("- info: %s", format)
	l.Logger.Printf(format, v...)

}

func (l *basicLogger) error(v ...interface{}) {
	l.Logger.Print("- error: ", fmt.Sprintln(v...))
}

func (l *basicLogger) errorf(format string, v ...interface{}) {
	format = fmt.Sprintf("- error: %s", format)
	l.Logger.Printf(format, v...)
}
