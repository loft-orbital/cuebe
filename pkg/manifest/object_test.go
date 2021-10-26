package manifest

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"github.com/tj/assert"
)

func TestDecode(t *testing.T) {
	// Nominal case
	raw := `
kind: "ConfigMap"
apiVersion: "v1"
metadata: name: "my-cm"
`
	ctx := cuecontext.New()
	m := New(ctx.CompileString(raw))

	obj, err := m.Decode()
	assert.NoError(t, err)
	gvk := obj.GetObjectKind().GroupVersionKind()
	assert.Equal(t, "v1", gvk.Version)
	assert.Equal(t, "", gvk.Group)
	assert.Equal(t, "ConfigMap", gvk.Kind)

	// Incomplete value
	raw = `
kind: string
apiVersion: "v1"
metadata: name: "my-cm"
`
	m = New(ctx.CompileString(raw))
	obj, err = m.Decode()
	assert.Error(t, err)
	assert.Regexp(t, "incomplete value", err)

	// Non valid manifest
	raw = `
apiVersion: "v1"
metadata: name: "my-cm"
`

	m = New(ctx.CompileString(raw))
	obj, err = m.Decode()
	assert.EqualError(t, err, "decoding manifest: Object 'Kind' is missing in '{\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"my-cm\"}}'")
}
