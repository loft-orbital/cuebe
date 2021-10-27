package manifest

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

type TypeMeta struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
}

// Extract extract kubernetes manifests from a cue.Value
func Extract(v cue.Value) []Manifest {
	manifests := []Manifest{}
	ctx := cuecontext.New()
	kmanifest := ctx.EncodeType(TypeMeta{})

	v.Walk(func(v cue.Value) bool {
		if kmanifest.Subsumes(v) {
			manifests = append(manifests, New(v))
			return false
		}
		return true
	}, nil)

	return manifests
}
