package release

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

type TypeMeta struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
}

// ExtractManifests extract kubernetes manifests from a cue.Value
func ExtractManifests(v cue.Value) []cue.Value {
	manifests := []cue.Value{}
	ctx := cuecontext.New()
	kmanifest := ctx.EncodeType(TypeMeta{})

	v.Walk(func(v cue.Value) bool {
		if kmanifest.Subsumes(v) {
			manifests = append(manifests, v)
			return false
		}
		return true
	}, nil)

	return manifests
}
