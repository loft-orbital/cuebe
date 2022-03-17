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
	"log"
	"sync"

	"cuelang.org/go/cue"
	"github.com/hashicorp/go-multierror"
	"github.com/loft-orbital/cuebe/pkg/manifest"
)

// Extract extracts all Manifests recursively starting from every paths.
// Recursion is stopped on `ignore` cue.Attribute or when a Manifest has been decoded.
// That means nested Manifests are not possible.
func Extract(v cue.Value, paths ...cue.Path) ([]manifest.Manifest, error) {
	res := make(chan interface{})
	var wg sync.WaitGroup

	// start from every paths
	for _, p := range paths {
		node := v.LookupPath(p)

		// walk value
		node.Walk(func(v cue.Value) bool {
			if a := v.Attribute("ignore"); a.Err() == nil {
				return false // stop diving, we've been told to
			}
			if manifest.IsManifest(v) {
				wg.Add(1)
				// extract manifest in goroutines
				go func(m cue.Value) {
					extract(m, res)
					wg.Done()
				}(v)
				return false // stop diving, we found a manifest
			}
			return true // continue deeper in this node
		}, nil)

	}

	// close chan when every extract are done
	go func() {
		wg.Wait()
		close(res)
	}()

	return collect(res)
}

func extract(v cue.Value, res chan<- interface{}) {
	m, err := manifest.Decode(v)
	if err != nil {
		res <- fmt.Errorf("failed to decode manifest at %s: %w", v.Path(), err)
		return
	}
	res <- m
}

func collect(res <-chan interface{}) (manifests []manifest.Manifest, err error) {
	for moe := range res {
		switch v := moe.(type) {
		case manifest.Manifest:
			manifests = append(manifests, v)
		case error:
			err = multierror.Append(err, v)
		default:
			log.Panicf("Unexpected manifest type: %T\n", v)
		}
	}

	return
}
