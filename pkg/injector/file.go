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
	"fmt"
	"io/fs"
	"path"

	"cuelang.org/go/cue"
	"github.com/PaesslerAG/jsonpath"
	"github.com/loft-orbital/cuebe/pkg/unifier"
	"gopkg.in/yaml.v3"
)

// File injector uses a local file as source of the inject value.
type File struct {
	path   cue.Path
	result chan interface{}
}

// NewFile creates a new file injector.
func NewFile(src, srcPath string, dstPath cue.Path, fs fs.FS) *File {
	r := make(chan interface{}, 1)
	go parseFile(src, srcPath, fs, r)
	return &File{path: dstPath, result: r}
}

// Inject returns the target value after injection.
func (f *File) Inject(target cue.Value) cue.Value {
	r := <-f.result
	return target.FillPath(f.path, r)
}

func parseFile(file, jpath string, fs fs.FS, res chan<- interface{}) {
	defer close(res)

	// read
	b, err := unifier.ReadFile(file, fs)
	if err != nil {
		res <- fmt.Errorf("failed to read %s: %w", file, err)
		return
	}

	var v interface{}

	if jpath == "" {
		// plain text injection
		v = string(b)
	} else {
		// structured injection
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

		v, err = jsonpath.Get(jpath, v)
		if err != nil {
			res <- fmt.Errorf("failed to extract path: %w", err)
			return
		}
	}

	res <- v
}
