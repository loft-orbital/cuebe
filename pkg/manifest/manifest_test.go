package manifest

import (
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const testManifestNominal = `
kind: "ConfigMap"
apiVersion: "v1"
metadata: name: "my-cm"
`

const testManifestIncomplete = `
kind: string
apiVersion: "v1"
metadata: name: "my-cm"
`

const testManifestNonValid = `
apiVersion: "v1"
metadata: name: "my-cm"
`

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
	m, err := Decode(ctx.CompileString(testManifestNominal))
	assert.NoError(t, err)
	assert.Equal(t, "my-cm", m.GetName())

	// Incomplete value
	m, err = Decode(ctx.CompileString(testManifestIncomplete))
	assert.Error(t, err)
	assert.Regexp(t, "incomplete value", err)

	// Non valid manifest
	m, err = Decode(ctx.CompileString(testManifestNonValid))
	assert.EqualError(t, err, "decoding manifest: Object 'Kind' is missing in '{\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"my-cm\"}}'")
}

func TestIsManifest(t *testing.T) {
	ctx := cuecontext.New()
	tests := map[string]struct {
		v        cue.Value
		expected bool
	}{
		"nominal": {
			v:        ctx.CompileString(testManifestNominal),
			expected: true,
		},
		"incomplete": {
			v:        ctx.CompileString(testManifestIncomplete),
			expected: true,
		},
		"nonValid": {
			v:        ctx.CompileString(testManifestNonValid),
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsManifest(tc.v))
		})
	}
}
