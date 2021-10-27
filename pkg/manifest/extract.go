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
package manifest

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

type TypeMeta struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
}

// Extract extract kubernetes manifests from a cue.Value
func Extract(v cue.Value) []Manifest {
	manifests := []Manifest{}
	ctx := cuecontext.New()
	kmanifest := ctx.EncodeType(TypeMeta{})

	v.Walk(func(v cue.Value) bool {
		if kmanifest.Subsumes(v) {
			manifests = append(manifests, New(v))
			return false
		}
		return true
	}, nil)

	return manifests
}
