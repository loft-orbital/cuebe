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
package modfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/module"
)

func TestParse(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "module.cue")
	// File does not exist
	_, err := Parse("")
	assert.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist)

	// empty file
	f, err := os.Create(filename)
	require.NoError(t, err)
	mf, err := Parse(filename)
	assert.NoError(t, err)
	assert.Equal(t, &File{}, mf)

	// cannot decode
	_, err = f.WriteString("module: 2")
	require.NoError(t, err)
	_, err = Parse(filename)
	assert.Error(t, err)
	require.NoError(t, f.Truncate(0))
	_, err = f.Seek(0, 0)
	require.NoError(t, err)

	// correct file
	_, err = f.WriteString(`
module: "host.com/potato/fries"

require: [
{path: "host.com/tomato/ketchup", version: "v1.2.3"},
{path: "host.com/spices/salt", version: "v0.1.2"},
{path: "host.com/lotof/fat/v2", version: "v2.1.5"},
]
`)
	require.NoError(t, err)
	mf, err = Parse(filename)
	assert.NoError(t, err)
	assert.Equal(t, "host.com/potato/fries", mf.Module)
	assert.ElementsMatch(t, []module.Version{
		{Path: "host.com/tomato/ketchup", Version: "v1.2.3"},
		{Path: "host.com/spices/salt", Version: "v0.1.2"},
		{Path: "host.com/lotof/fat/v2", Version: "v2.1.5"},
	},
		mf.Require)

	// build error
	_, err = f.WriteString(`
module: int
module: "potato"
`)
	require.NoError(t, err)
	_, err = Parse(filename)
	assert.Error(t, err)
}
