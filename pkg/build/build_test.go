package build

import (
	"testing"

	"github.com/loft-orbital/cuebe/pkg/context"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestBuild(t *testing.T) {
	// build a context
	bctx := context.New()
	fsys := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fsys, "main.cue",
		[]byte("package main\nhello: string @inject(type=file, src=inject.yml, path=$.name)"),
		0666,
	))
	require.NoError(t, afero.WriteFile(fsys, "inject.yml", []byte("name: cuebe"), 0666))
	require.NoError(t, bctx.Add(fsys))

	v, err := Build(bctx, nil)
	assert.NoError(t, err)
	name, err := v.Lookup("hello").String()
	assert.NoError(t, err)
	assert.Equal(t, "cuebe", name)

	// add a bit of non-sense
	require.NoError(t, afero.WriteFile(fsys, "error.cue", []byte("package main\nhello: 42"), 0666))
	require.NoError(t, bctx.Add(fsys))

	v, err = Build(bctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicting values 42 and string")
}
