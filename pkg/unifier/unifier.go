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
	"fmt"
	"io/fs"
	"path"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

// Unifier represents a multi sources CUE release.
type Unifier struct {
	ctx *cue.Context

	vLock  sync.RWMutex
	values []cue.Value
}

// Load loads instances inside entrypoints, using wd as a working directory.
// It returns the Unifier containing all the values find.
func Load(entrypoints []string, cfg *load.Config) (*Unifier, error) {
	u := &Unifier{
		ctx: cuecontext.New(),
	}
	bis := load.Instances(entrypoints, cfg)
	var err error
	u.values, err = u.ctx.BuildInstances(bis)
	if err != nil {
		return nil, fmt.Errorf("failed to build instances: %w", err)
	}

	return u, nil
}

// Unify reduces all the Unifier values in a single cue.Value
func (u *Unifier) Unify() cue.Value {
	u.vLock.RLock()
	defer u.vLock.RUnlock()

	if len(u.values) <= 0 {
		return cue.Value{}
	}
	v := u.values[0]
	for _, nv := range u.values[1:] {
		v = v.Unify(nv)
	}

	return v
}

// AddFile parse and compile an orphan file, then add it to the Unifier's values.
// It plain texti (cue,yaml,json) or sops-encrypted files.
func (u *Unifier) AddFile(file string, fsys fs.FS) error {
	um, err := UnmarshallerFor(path.Ext(file))
	if err != nil {
		return fmt.Errorf("failed to add %s: %w", file, err)
	}

	b, err := ReadFile(file, fsys)
	if err != nil {
		return fmt.Errorf("failed to add %s: %w", file, err)
	}

	v, err := um.Unmarshal(b, u.ctx, cue.Filename(file))
	if err != nil {
		return fmt.Errorf("failed to add %s: %w", file, err)
	}

	u.vLock.Lock()
	defer u.vLock.Unlock()
	u.values = append(u.values, v)
	return nil
}
