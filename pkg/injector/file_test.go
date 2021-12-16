package injector

import (
	"encoding/json"
	"io/ioutil"
	"os"
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
	f, err := ioutil.TempFile("", "test_inject*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	b, _ := json.Marshal(Spacecraft{"Voyager", 470, 1})
	f.Write(b)
	res := make(chan interface{})
	go parseFile(f.Name(), "$.power", res)

	r := <-res
	assert.Equal(t, float64(470), r)
}

func TestInject(t *testing.T) {
	f, err := ioutil.TempFile("", "test_inject*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	ctx := cuecontext.New()
	v := ctx.CompileString("spacecraft: name: string")

	b, _ := json.Marshal(Spacecraft{"Voyager", 470, 1})
	f.Write(b)
	fi := NewFile(f.Name(), "$.name", cue.ParsePath("spacecraft.name"))
	v = fi.Inject(v)

	actual, err := v.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `{"spacecraft":{"name":"Voyager"}}`, string(actual))
}
