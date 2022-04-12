package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestManifestId(t *testing.T) {
	u := new(unstructured.Unstructured)
	u.SetKind("Deployment")
	u.SetAPIVersion("apps/v1")
	u.SetName("potato")
	u.SetNamespace("tomato")
	m := New(u)

	expected := Id{
		Kind:      "Deployment",
		Version:   "v1",
		Group:     "apps",
		Namespace: "tomato",
		Name:      "potato",
	}
	assert.Equal(t, expected, m.Id())
}
