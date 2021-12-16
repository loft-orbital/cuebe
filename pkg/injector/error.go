package injector

import (
	"fmt"

	"cuelang.org/go/cue"
)

type Error struct {
	path cue.Path
	err  error
}

func NewError(err error, dstPath cue.Path) *Error {
	return &Error{path: dstPath, err: fmt.Errorf("injection error: %w", err)}
}

func (e *Error) Inject(target cue.Value) cue.Value {
	return target.FillPath(e.path, e.err)
}
