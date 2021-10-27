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

	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
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
