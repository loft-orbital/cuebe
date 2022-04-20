package context

import (
	iofs "io/fs"
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFS(t *testing.T) {
	ctx := New()
	assert.Implements(t, (*iofs.FS)(nil), ctx.GetFS())
}

func TestCopyFile(t *testing.T) {
	payload := []byte("Hello cuebe!")
	filename := "foo.bar"
	src := afero.NewMemMapFs()
	dst := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(src, filename, payload, 0666))

	written, err := CopyFile(dst, src, filename, 0666)
	assert.NoError(t, err)
	assert.Equal(t, len(payload), int(written))

	// File now exists, test overwrite
	payload = []byte("!ebeuc olleH")
	require.NoError(t, afero.WriteFile(src, filename, payload, 0666))

	written, err = CopyFile(dst, src, filename, 0666)
	assert.NoError(t, err)
	actual, err := afero.ReadFile(dst, filename)
	assert.NoError(t, err)
	assert.Equal(t, payload, actual)
}

func TestContextAdd(t *testing.T) {
	// Prepare source
	src := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(src, "cue.mod/module.cue", []byte("module: \"github.com/super/module\""), 0666))
	require.NoError(t, afero.WriteFile(src, "dir/nested/file", []byte("dir/nested/file"), 0666))
	require.NoError(t, afero.WriteFile(src, "dir/file", []byte("dir/file"), 0666))
	require.NoError(t, afero.WriteFile(src, "file", []byte("file"), 0666))
	require.NoError(t, src.MkdirAll("empty/dir", 0775))

	ctx := New()
	assert.NoError(t, ctx.Add(src))
	cfs := ctx.GetFS()
	f, err := cfs.Open("dir/nested/file")
	assert.NoError(t, err)
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	assert.NoError(t, err)
	assert.Equal(t, []byte("dir/nested/file"), data)
}

func TestCopy(t *testing.T) {
	// Prepare source
	src := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(src, "cue.mod/module.cue", []byte("module: \"github.com/super/module\""), 0666))
	require.NoError(t, afero.WriteFile(src, "dir/nested/file", []byte("dir/nested/file"), 0666))
	require.NoError(t, afero.WriteFile(src, "dir/file", []byte("dir/file"), 0666))
	require.NoError(t, afero.WriteFile(src, "file", []byte("file"), 0666))
	require.NoError(t, src.MkdirAll("empty/dir", 0775))

	// test
	dst := afero.NewMemMapFs()
	assert.NoError(t, Copy(dst, src))
	stat, err := dst.Stat("dir/nested")
	assert.NoError(t, err)
	assert.True(t, stat.IsDir(), "'dir/nested' should be a directory")
	stat, err = dst.Stat("cue.mod/module.cue")
	assert.NoError(t, err)
	assert.False(t, stat.IsDir(), "'cue.mod/module.cue' should be a file")
	assert.Equal(t, len([]byte("module: \"github.com/super/module\"")), int(stat.Size()))
}
