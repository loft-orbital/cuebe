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
package build

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"github.com/loft-orbital/cuebe/pkg/context"
	"github.com/loft-orbital/cuebe/pkg/injector"
	"github.com/loft-orbital/cuebe/pkg/unifier"
	"github.com/spf13/afero"
)

// Build builds a context into a single cue.Value,
// performing all `cuebe` flavored features.
func Build(bctx *context.Context, cfg *load.Config) (cue.Value, error) {
	// NOTE: this local copy is done until cue itself support loading from a fs.FS.
	// c.f. https://github.com/cue-lang/cue/issues/607
	tempdir, err := afero.TempDir(afero.NewOsFs(), "", "cuebe-build-")
	if err != nil {
		return cue.Value{}, fmt.Errorf("could not create temp directory: %w", err)
	}
	tempfs := afero.NewBasePathFs(afero.NewOsFs(), tempdir)
	defer tempfs.RemoveAll("")
	if err := context.Copy(tempfs, bctx.GetAferoFS()); err != nil {
		return cue.Value{}, fmt.Errorf("could not copy context to temp directory: %w", err)
	}

	// overwrite load config
	if cfg == nil {
		cfg = new(load.Config)
	}
	cfg.Dir = tempdir

	// load context
	u, err := unifier.Load([]string{}, cfg)
	if err != nil {
		return cue.Value{}, fmt.Errorf("failed to load context: %w", err)
	}
	v := u.Unify()

	// do injections
	v = injector.Inject(v, bctx.GetFS())

	return v, v.Err()
}
