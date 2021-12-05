package injector

import "cuelang.org/go/cue"

type Injector interface {
	Inject(v cue.Value) cue.Value
}
