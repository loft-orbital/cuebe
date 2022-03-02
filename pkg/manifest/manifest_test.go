package manifest

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNew(t *testing.T) {
	u := new(unstructured.Unstructured)
	u.SetAPIVersion("v1")
	u.SetKind("Namespace")
	u.SetName("potato")

	m := New(u)
	assert.Equal(t, Manifest{u}, m)
}

func TestDecode(t *testing.T) {
	ctx := cuecontext.New()

	// Nominal case
	raw := `
kind: "ConfigMap"
apiVersion: "v1"
metadata: name: "my-cm"
`
	m, err := Decode(ctx.CompileString(raw))
	assert.NoError(t, err)
	assert.Equal(t, "my-cm", m.GetName())

	// Incomplete value
	raw = `
kind: string
apiVersion: "v1"
metadata: name: "my-cm"
`
	m, err = Decode(ctx.CompileString(raw))
	assert.Error(t, err)
	assert.Regexp(t, "incomplete value", err)

	// Non valid manifest
	raw = `
apiVersion: "v1"
metadata: name: "my-cm"
`
	m, err = Decode(ctx.CompileString(raw))
	assert.EqualError(t, err, "decoding manifest: Object 'Kind' is missing in '{\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"my-cm\"}}'")
}
