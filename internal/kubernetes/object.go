package kubernetes

import (
	"k8s.io/apimachinery/pkg/runtime"
)

var primaryKind = map[string]bool{
	"Namespace":                true,
	"CustomResourceDefinition": true,
}

// SplitObj splits kubernetes object into two distinct list based on their kind.
// The primary list will contains object that need to be deployed first.
func SplitObj(objs []runtime.Object) (primary []runtime.Object, secondary []runtime.Object) {
	for _, obj := range objs {
		kind := obj.GetObjectKind().GroupVersionKind().Kind
		if _, isPrimary := primaryKind[kind]; isPrimary {
			primary = append(primary, obj)
		} else {
			secondary = append(secondary, obj)
		}
	}

	return primary, secondary
}
