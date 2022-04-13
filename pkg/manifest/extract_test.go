/*
Copyright © 2021 Loft Orbital

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package manifest

import (
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testError struct{}

func (te *testError) Error() string {
	return "A test error"
}

func TestExtract(t *testing.T) {
	ctx := cuecontext.New()
	v := ctx.CompileString(`
manifest: {
	apiVersion: "v1"
	kind:       "ConfigMap"
	metadata: name: "notinpath"
	data: foo:      24
}

path1: {
	manifesti: {
		apiVersion: "v1"
		kind:       "ConfigMap"
		metadata: name: "ignore"
		data: foo:      24
	} @ignore()

	manifest: {
		apiVersion: "v1"
		kind:       "ConfigMap"
		metadata: name: "path1"
		data: foo:      42
	}
}

path2: {
	manifeste: {
		apiVersion: "v1"
		kind:       string
		metadata: name: "error"
		data: foo:      24
	}

	manifest: {
		apiVersion: "v1"
		kind:       "ConfigMap"
		metadata: name: "path2"
		data: foo:      42
	}
}
`)
	require.NoError(t, v.Err())

	mfs, err := Extract(v, cue.ParsePath("path1"), cue.ParsePath("path2"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode manifest at path2.manifeste")
	assert.Len(t, mfs, 2)
	names := []string{mfs[0].GetName(), mfs[1].GetName()}
	assert.Contains(t, names, "path1")
	assert.Contains(t, names, "path2")
}

func TestCollect(t *testing.T) {
	err1 := new(testError)
	err2 := new(testError)

	res := make(chan interface{}, 5)
	res <- New(nil)
	res <- New(nil)
	res <- New(nil)
	res <- err1
	res <- err2
	close(res)

	mfs, err := collect(res)
	assert.Len(t, mfs, 3)
	assert.ErrorAs(t, err, &err1)
	assert.ErrorAs(t, err, &err2)

	// panic
	res = make(chan interface{}, 1)
	res <- 1
	assert.Panics(t, func() { collect(res) })
}

func BenchmarkExtractManifest(b *testing.B) {
	input := `{
			  root: {kind: "Secret", apiVersion: "v1"}
			  foo: {bar: "baz"}
			  nested: list: [{kind: "ConfigMap", apiVersion: "v1"}, {kind: "Deployment", apiVersion: "apps/v1"}, {foo: "bar"}]
			  }`
	ctx := cuecontext.New()
	v := ctx.CompileString(input)
	mfs, err := Extract(v, cue.ParsePath("."))
	assert.NoError(b, err)
	assert.Len(b, mfs, 3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Extract(v, cue.ParsePath("."))
	}
}
