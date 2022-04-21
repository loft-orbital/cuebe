package log

import "context"

type ctxKey struct{}

func WithLogger(parent context.Context, logger Logger) context.Context {
	return context.WithValue(parent, ctxKey{}, logger)
}

func GetLogger(ctx context.Context) Logger {
	v := ctx.Value(ctxKey{})
	log, ok := v.(Logger)
	if !ok {
		return DiscardLogger
	}
	return log
}
