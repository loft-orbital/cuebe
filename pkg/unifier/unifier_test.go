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
package unifier

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const raw = `
package main

hello: "cuebe"
`

func TestLoad(t *testing.T) {
	d, err := ioutil.TempDir("", "cuebetest")
	require.NoError(t, err)
	defer os.RemoveAll(d)
	f, err := ioutil.TempFile(d, "*.cue")
	require.NoError(t, err)
	_, err = f.WriteString(raw)
	require.NoError(t, err)
	f2, err := ioutil.TempFile(d, "*.cue")
	require.NoError(t, err)
	_, err = f2.WriteString("hello: false")
	require.NoError(t, err)

	// Empty entypoints
	u, err := Load([]string{}, &load.Config{Dir: d})
	require.NoError(t, err)
	require.Len(t, u.values, 1)
	b, _ := u.values[0].MarshalJSON()
	assert.Equal(t, "{\"hello\":\"cuebe\"}", string(b))

	// Cue error (both files)
	u, err = Load([]string{f2.Name(), f.Name()}, &load.Config{Dir: d})
	assert.EqualError(t, err, "failed to build instances: hello: conflicting values false and \"cuebe\" (mismatched types bool and string)")

	d2, err := ioutil.TempDir("", "cuebetest")
	require.NoError(t, err)
	defer os.RemoveAll(d2)
	// Empty both
	u, err = Load([]string{}, &load.Config{Dir: d2})
	assert.EqualError(t, err, "failed to load instances: no CUE files in .")

	// Empty directory
	u, err = Load([]string{f2.Name()}, nil)
	require.NoError(t, err)
	require.Len(t, u.values, 1)
	b, _ = u.values[0].MarshalJSON()
	assert.Equal(t, "{\"hello\":false}", string(b))
}

func TestUnify(t *testing.T) {
	// No values
	u := &Unifier{}
	assert.Equal(t, cue.Value{}, u.Unify())

	// 1 value
	ctx := cuecontext.New()
	u.values = []cue.Value{ctx.CompileString(raw)}
	assert.Equal(t, ctx.CompileString(raw), u.Unify())

	// Multiple values
	u.values = append(u.values, ctx.CompileString("foo:true"))
	b, err := u.Unify().MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, "{\"hello\":\"cuebe\",\"foo\":true}", string(b))
}

func TestAddFile(t *testing.T) {
	u := &Unifier{ctx: cuecontext.New()}

	// Unsupported extension
	assert.EqualError(t, u.AddFile("file.unsupported"), "failed to add file.unsupported: Unsupported extension .unsupported")

	// Bad file
	if runtime.GOOS == "windows" {
		assert.EqualError(t, u.AddFile("file.yaml"), "failed to add file.yaml: open file.yaml: The system cannot find the file specified.")
	} else {
		assert.EqualError(t, u.AddFile("file.yaml"), "failed to add file.yaml: open file.yaml: no such file or directory")
	}

	// Failed unmarshal
	f, err := ioutil.TempFile("", "*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString("bad")
	assert.EqualError(t, u.AddFile(f.Name()), fmt.Sprintf("failed to add %s: failed to unmarshal json: json: invalid JSON", f.Name()))

	// Nominal case
	f, err = ioutil.TempFile("", "*.cue")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString("foo: true")
	assert.NoError(t, u.AddFile(f.Name()))
	require.Len(t, u.values, 1)
	b, err := u.values[0].MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, "{\"foo\":true}", string(b))
}
