package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	rt "k8s.io/apimachinery/pkg/runtime/testing"
)

func TestSplitObj(t *testing.T) {
	objs := []runtime.Object{
		&rt.ObjectTest{TypeMeta: runtime.TypeMeta{APIVersion: "v1", Kind: "Namespace"}},
		&rt.ObjectTest{TypeMeta: runtime.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"}},
		&rt.ObjectTest{TypeMeta: runtime.TypeMeta{APIVersion: "v1", Kind: "CustomResourceDefinition"}},
	}

	p, s := SplitObj(objs)
	assert.ElementsMatch(t, p, []runtime.Object{objs[0], objs[2]})
	assert.ElementsMatch(t, s, []runtime.Object{objs[1]})
}
