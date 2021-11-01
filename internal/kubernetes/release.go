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
package kubernetes

import (
	"context"
	"fmt"
	"sort"

	"cuelang.org/go/cue"
	"github.com/loft-orbital/cuebe/pkg/manifest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type Release struct {
	cfg *rest.Config

	// Objects managed by this release
	Objects []unstructured.Unstructured
}

// Create a new release by extracting manifests and context from a CUE value.
// If ktxpath fails to resolve to a string, ktxfallback will be used as the Kubernetes context.
func NewReleaseFor(v cue.Value, ktxpath string, ktxfallback string) (*Release, error) {
	// Get kubernetes client
	ktx, err := ExtractContext(v, ktxpath)
	if err != nil {
		ktx = ktxfallback
	}
	cfg, err := DefaultConfig(ktx)
	if err != nil {
		return nil, fmt.Errorf("failed to build release: %w", err)
	}

	// Get manifests
	mfs := manifest.Extract(v)
	objs := make([]unstructured.Unstructured, len(mfs))
	for i, m := range mfs {
		objs[i], err = m.ToObj()
		if err != nil {
			return nil, fmt.Errorf("failed to build release: %w", err)
		}
	}
	sort.Sort(SortableUnstructured(objs))

	return &Release{
		cfg:     cfg,
		Objects: objs,
	}, nil
}

// Deploy patches every objects contained in the release to the Kubernetes cluster.
// It uses server-side apply method.
func (r *Release) Deploy(ctx context.Context) error {
	// Get dynamic client
	dync, err := dynamic.NewForConfig(r.cfg)
	if err != nil {
		return err
	}

	// Get discovery client
	disc, err := discovery.NewDiscoveryClientForConfig(r.cfg)
	if err != nil {
		return err
	}
	// Get rest mapper
	rm := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disc))

	todo := make([]unstructured.Unstructured, len(r.Objects))
	copy(todo, r.Objects)
	failed := []unstructured.Unstructured{}
	// Start patching
	for len(todo) > 0 {
		rm.Reset() // Invalidate the cache to get new resource mappings

		for _, obj := range todo {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				dr, err := dynamicResourceInterfaceFor(obj, rm, dync)
				if err != nil {
					failed = append(failed, obj) // Couldn't find any mapping, retrying next loop
				}

				data, err := obj.MarshalJSON()
				if err != nil {
					return fmt.Errorf("Couldn't marshal %s: %w", printObj(obj), err)
				}

				// TODO maybe retrieve from option, to enable dryrun
				po := metav1.PatchOptions{
					FieldManager: "cuebe",
				}
				// Patch object
				if _, err := dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, po); err != nil {
					return fmt.Errorf("Couldn't patch %s: %w", printObj(obj), err)
				}
				fmt.Printf("%s patched\n", printObj(obj))
			}
		}

		// Found a deadlock
		if len(failed) >= len(todo) {
			return fmt.Errorf("Failed to patch every objects, %d remaining", len(failed))
		}
		todo = failed
		failed = []unstructured.Unstructured{}
	}

	return nil
}

func (r *Release) Host() string {
	return r.cfg.Host

}
