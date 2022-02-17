package gen

import (
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/loft-orbital/cuebe/pkg/modfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectGoPkg(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip windows until we fix the collect function")
	}

	// create a temp workdir
	wd, err := ioutil.TempDir("", "cuebe-gen-test")
	require.NoError(t, err)
	defer os.RemoveAll(wd)
	// prepare directories
	require.NoError(t, os.MkdirAll(path.Join(wd, "cue.mod", "pkg", "host.com", "foo", "bar", "cue.mod"), 0700))
	require.NoError(t, os.MkdirAll(path.Join(wd, "cue.mod", "gen", "host.com", "foo", "bar", "cue.mod"), 0700))
	// add some cue.mod files
	require.NoError(t, os.WriteFile(
		path.Join(wd, modfile.CueModFile),
		[]byte("module: \"host.com/potato/chips\"\ngodef: [\"host.com/sauce/ketchup\"]"),
		0666))
	require.NoError(t, os.WriteFile(
		path.Join(wd, "cue.mod", "pkg", "host.com", "foo", "bar", modfile.CueModFile),
		[]byte("module: \"host.com/potato/chips\"\ngodef: [\"host.com/type/yam\"]"),
		0666))
	require.NoError(t, os.WriteFile(
		path.Join(wd, "cue.mod", "gen", "host.com", "foo", "bar", modfile.CueModFile),
		[]byte("module: \"host.com/potato/chips\"\ngodef: [\"host.com/type/sweet\"]"),
		0666))

	pkgs, err := CollectGoPkg(wd)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"host.com/sauce/ketchup", "host.com/type/yam"}, pkgs)
}
