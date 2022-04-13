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
	"crypto/sha1"
	"fmt"

	"cuelang.org/go/cue"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Manifest is a wrapper around *unstructured.Unstructured.
type Manifest struct {
	*unstructured.Unstructured
}

// Hash returns the hash of a Manifest.
func (m Manifest) Hash() ([]byte, error) {
	data, err := m.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("could not json marshal: %w", err)
	}
	sha := sha1.Sum(data)

	return sha[:], nil
}

// New creates a new Manifest from an *unstructured.Unstructured object.
func New(u *unstructured.Unstructured) Manifest {
	return Manifest{u}
}

// Decode converts a cue.Value into a Manifest
// or returns an error if the value is not compatible with a k8s object.
func Decode(v cue.Value) (Manifest, error) {
	unstruct := new(unstructured.Unstructured)
	if err := v.Decode(unstruct); err != nil {
		return Manifest{unstruct}, fmt.Errorf("decoding manifest: %w", err)
	}
	return Manifest{unstruct}, nil
}

// IsManifest returns true if the cue.Value "looks like" a Manifest.
// For now that means it has a 'kind' and 'apiVersion', both being strings.
func IsManifest(v cue.Value) bool {
	k, vs := v.LookupPath(cue.MakePath(cue.Str("kind"))), v.LookupPath(cue.MakePath(cue.Str("apiVersion")))
	return k.IncompleteKind() == cue.StringKind && vs.IncompleteKind() == cue.StringKind
}

// IsRemote returns true if the manifest has an uuid set.
// TODO: this is weak, get rid of using that asap.
func (m Manifest) IsRemote() bool {
	return string(m.GetUID()) != ""
}
