package manifest

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"github.com/tj/assert"
)

func TestExtractManifestt(t *testing.T) {
	tcs := map[string]struct {
		input string
		count int
	}{
		"list": {
			input: `[
			  {kind: "Deployment", apiVersion: "apps/v1"},
			  {kind: "Secret", apiVersion: "v1"},
			  {kind: "ConfigMap", apiVersion: "v1"},
			  {foo: "bar"},
			  ]`,
			count: 3,
		},
		"struct": {
			input: `{
			  secret: {kind: "Secret", apiVersion: "v1"}
			  cm: {kind: "ConfigMap", apiVersion: "v1"}
			  foo: {bar: "baz"}
			  }`,
			count: 2,
		},
		"complex": {
			input: `{
			  root: {kind: "Secret", apiVersion: "v1"}
			  foo: {bar: "baz"}
			  nested: list: [{kind: "ConfigMap", apiVersion: "v1"}, {kind: "Deployment", apiVersion: "apps/v1"}, {foo: "bar"}]
			  }`,
			count: 3,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := cuecontext.New()
			v := ctx.CompileString(tc.input)
			mfs := Extract(v)
			assert.Len(t, mfs, tc.count)
		})
	}
}

func BenchmarkExtractManifest(b *testing.B) {
	input := `{
			  root: {kind: "Secret", apiVersion: "v1"}
			  foo: {bar: "baz"}
			  nested: list: [{kind: "ConfigMap", apiVersion: "v1"}, {kind: "Deployment", apiVersion: "apps/v1"}, {foo: "bar"}]
			  }`
	ctx := cuecontext.New()
	v := ctx.CompileString(input)
	mfs := Extract(v)
	assert.Len(b, mfs, 3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Extract(v)
	}
}
