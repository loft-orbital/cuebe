package injector

import (
	"encoding/json"
	"fmt"
	"path"

	"cuelang.org/go/cue"
	"github.com/PaesslerAG/jsonpath"
	"github.com/loft-orbital/cuebe/pkg/unifier"
	"gopkg.in/yaml.v2"
)

type File struct {
	path   cue.Path
	result chan interface{}
}

func NewFile(src, srcPath, dstPath string) *File {
	r := make(chan interface{}, 1)
	go parseFile(src, srcPath, r)
	return &File{path: cue.ParsePath(dstPath), result: r}
}

func (f *File) Inject(target cue.Value) cue.Value {
	r := <-f.result
	return target.FillPath(f.path, r)
}

func parseFile(file, jpath string, res chan<- interface{}) {
	defer close(res)

	// read
	b, err := unifier.ReadFile(file)
	if err != nil {
		res <- err
		return
	}

	// unmarshal
	var v interface{}
	switch path.Ext(file) {
	case ".json":
		err = json.Unmarshal(b, &v)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(b, &v)
	default:
		err = fmt.Errorf("Unsupported extension %s", path.Ext(file))
	}
	if err != nil {
		res <- fmt.Errorf("failed to unmarshal: %w", err)
		return
	}

	// get path
	v, err = jsonpath.Get(jpath, v)
	if err != nil {
		res <- fmt.Errorf("failed to extract path: %w", err)
		return
	}

	res <- v
}
