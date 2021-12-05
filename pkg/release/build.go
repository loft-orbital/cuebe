package release

import (
	"cuelang.org/go/cue"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Build(v cue.Value) ([]unstructured.Unstructured, error) {
	mp := []cue.Path{}

	v.Walk(func(v cue.Value) bool {
		if a := v.Attribute("ignore"); a.Err() == nil {
			return false
		}

		// Check if could be a manifest
		k, vs := v.LookupPath(cue.MakePath(cue.Str("kind"))), v.LookupPath(cue.MakePath(cue.Str("apiVersion")))
		if k.Kind() == cue.StringKind && vs.Kind() == cue.StringKind {
			mp = append(mp, v.Path())
			return false
		}
		return true
	}, nil)
}
