package manifest

import "cuelang.org/go/cue"

type Manifest struct {
	cue.Value
}

func New(v cue.Value) Manifest {
	return Manifest{v}
}
