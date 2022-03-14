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
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFile(t *testing.T) {
	// Encrypted
	fe, err := ioutil.TempFile("", "*.enc.yaml")
	require.NoError(t, err)
	defer os.Remove(fe.Name())
	fsys := os.DirFS(path.Dir(fe.Name()))

	_, err = fe.WriteString("hello: 1")
	require.NoError(t, err)

	_, err = ReadFile(path.Base(fe.Name()), fsys)
	assert.EqualError(t, err, "could not decrypt data: sops metadata not found")

	// Plain text
	f, err := ioutil.TempFile("", "*.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	fsys = os.DirFS(path.Dir(f.Name()))

	_, err = f.WriteString("hello: 1")
	require.NoError(t, err)

	b, err := ReadFile(path.Base(f.Name()), fsys)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello: 1"), b)
}
