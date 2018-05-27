package log

import "os"

var l Logger

type Logger interface {
	info(v ...interface{})
	infof(format string, v ...interface{})
	error(v ...interface{})
	errorf(format string, v ...interface{})
}

func init() {
	InitBasicLogger(os.Stdout)
}

// Info log @info level
func Info(v ...interface{}) {
	l.info(v...)
}

// Infof -> printf
func Infof(format string, v ...interface{}) {
	l.infof(format, v...)
}

// Error log @error level
func Error(v ...interface{}) {
	l.error(v...)
}

// Errorf -> printf
func Errorf(format string, v ...interface{}) {
	l.errorf(format, v...)
}
