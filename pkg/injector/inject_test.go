package injector

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInjectNil(t *testing.T) {
	ctx := cuecontext.New()
	v := ctx.CompileString("foo: _ @inject(type=nil)")

	v = Inject(v, nil)
	assert.EqualError(t, v.Err(), "injection error: unsupported injector type nil")
}

func TestInjectNoType(t *testing.T) {
	ctx := cuecontext.New()
	v := ctx.CompileString("foo: _ @inject()")

	v = Inject(v, nil)
	assert.EqualError(t, v.Err(), "injection error: missing injector type")
}

func TestInjectFileNoSrc(t *testing.T) {
	ctx := cuecontext.New()
	v := ctx.CompileString("foo: _ @inject(type=file)")

	v = Inject(v, nil)
	assert.EqualError(t, v.Err(), "injection error: missing src key for file injector")
}

func TestInjectFile(t *testing.T) {
	f, err := ioutil.TempFile("", "*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()
	_, err = f.WriteString("{\"potato\": 42}")
	require.NoError(t, err)

	ctx := cuecontext.New()
	v := ctx.CompileString(fmt.Sprintf("foo: _ @inject(type=file, src=%s, path=$.potato)", path.Base(f.Name())))

	v = Inject(v, os.DirFS(path.Dir(f.Name())))
	assert.NoError(t, v.Err())
	json, err := v.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, "{\"foo\":42}", string(json))
}
