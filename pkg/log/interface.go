package log

type any = interface{}

type Logger interface {
	Info(format string, v ...any)
	Error(format string, v ...any)
}
