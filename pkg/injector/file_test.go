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
package injector

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Spacecraft struct {
	Name     string `json:"name"`
	Power    int    `json:"power"`
	Revision int    `json:"revision"`
}

func TestParseFile(t *testing.T) {
	if runtime.GOOS == "windows" && os.Getenv("CI") != "" {
		t.Skip("skipping fs related test on windows")
	}

	f, err := ioutil.TempFile("", "test_inject*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()
	fsys := os.DirFS(path.Dir(f.Name()))

	b, _ := json.Marshal(Spacecraft{"Voyager", 470, 1})
	f.Write(b)
	res := make(chan interface{})
	go parseFile(path.Base(f.Name()), "$.power", fsys, res)

	assert.Equal(t, float64(470), <-res)

	// plain
	f, err = ioutil.TempFile("", "test_inject*.plain")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()
	fsys = os.DirFS(path.Dir(f.Name()))

	f.WriteString("hello cuebe!")
	res = make(chan interface{})
	go parseFile(path.Base(f.Name()), "", fsys, res)

	// r = <-res
	assert.Equal(t, "hello cuebe!", <-res)
}

func TestInject(t *testing.T) {
	if runtime.GOOS == "windows" && os.Getenv("CI") != "" {
		t.Skip("skipping fs related test on windows")
	}

	f, err := ioutil.TempFile("", "test_inject*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()
	fsys := os.DirFS(path.Dir(f.Name()))

	ctx := cuecontext.New()
	v := ctx.CompileString("spacecraft: name: string")

	b, _ := json.Marshal(Spacecraft{"Voyager", 470, 1})
	f.Write(b)
	fi := NewFile(path.Base(f.Name()), "$.name", cue.ParsePath("spacecraft.name"), fsys)
	v = fi.Inject(v)

	actual, err := v.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `{"spacecraft":{"name":"Voyager"}}`, string(actual))
}
