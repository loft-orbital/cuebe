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
	"fmt"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/parser"
)

// Parse parses a CUE modfile (often located in cue.mod/module.cue).
func Parse(file string) (*File, error) {
	cf, err := parser.ParseFile(file, nil)
	if err != nil {
		return nil, fmt.Errorf("could not parse modfile: %w", err)
	}
	ctx := cuecontext.New()
	v := ctx.BuildFile(cf)
	if v.Err() != nil {
		return nil, fmt.Errorf("could not build modfile: %w", v.Err())
	}
	f := &File{}
	if err := v.Decode(f); err != nil {
		return nil, fmt.Errorf("failed to decode modfile: %w", err)
	}
	return f, nil
}
