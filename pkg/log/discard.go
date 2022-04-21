package log

type discardLogger struct{}

var DiscardLogger = &discardLogger{}

func (l *discardLogger) Info(format string, v ...any)  {}
func (l *discardLogger) Error(format string, v ...any) {}
